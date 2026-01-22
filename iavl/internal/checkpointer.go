package internal

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/tidwall/btree"
)

type Checkpointer struct {
	lastCheckpoint         atomic.Uint32
	reqChan                chan checkpointReq
	doneChan               chan error
	changesetLock          sync.RWMutex
	changesetsByCheckpoint *btree.Map[uint32, *Changeset]
	evictor                Evictor
}

func NewCheckpointer(evictor Evictor) *Checkpointer {
	cp := &Checkpointer{
		reqChan:                make(chan checkpointReq, 16),
		doneChan:               make(chan error, 1),
		changesetsByCheckpoint: &btree.Map[uint32, *Changeset]{},
		evictor:                evictor,
	}
	cp.start()
	return cp
}

func (cp *Checkpointer) start() {
	go func() {
		err := cp.proc()
		cp.doneChan <- err
	}()
}

// ChangesetByCheckpoint finds the changeset containing the given checkpoint.
// Since checkpoint == version, this works for both checkpoint and version lookups.
func (cp *Checkpointer) ChangesetByCheckpoint(checkpoint uint32) *Changeset {
	cp.changesetLock.RLock()
	defer cp.changesetLock.RUnlock()

	var res *Changeset
	// Find the changeset with the highest start checkpoint <= the requested checkpoint
	cp.changesetsByCheckpoint.Descend(checkpoint, func(key uint32, cs *Changeset) bool {
		res = cs
		return false // Take the first (highest) entry <= checkpoint
	})
	return res
}

func (cp *Checkpointer) Checkpoint(writer *ChangesetWriter, root *NodePointer, version uint32, seal bool) error {
	select {
	case err := <-cp.doneChan:
		return err
	default:
	}
	cp.reqChan <- checkpointReq{
		writer:  writer,
		root:    root,
		version: version,
		seal:    seal,
	}
	return nil
}

func (cp *Checkpointer) proc() error {
	var curWriter *ChangesetWriter
	for req := range cp.reqChan {
		_, span := Tracer.Start(context.Background(), "SaveCheckpoint")

		// checkpoint == version
		if err := req.writer.SaveCheckpoint(req.version, req.root); err != nil {
			return err
		}
		if req.seal {
			if err := req.writer.Seal(req.version); err != nil {
				return err
			}
		} else {
			if err := req.writer.CreatedSharedReader(); err != nil {
				return err
			}
		}
		// if we have a new writer, update the changeset map
		// we only need to store the changeset ONCE per writer for the FIRST checkpoint it writes
		if req.writer != curWriter { // compare pointers
			curWriter = req.writer
			cp.changesetLock.Lock()
			cp.changesetsByCheckpoint.Set(req.version, curWriter.Changeset())
			cp.changesetLock.Unlock()
		}
		cp.lastCheckpoint.Store(req.version)
		// notify evictor
		if cp.evictor != nil {
			cp.evictor.Evict(req.root, req.version)
		}

		span.End()
	}
	return nil
}

func (cp *Checkpointer) Close() error {
	close(cp.reqChan)
	return <-cp.doneChan
}

type checkpointReq struct {
	writer  *ChangesetWriter
	root    *NodePointer
	version uint32
	seal    bool
}
