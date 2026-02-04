package internal

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/tidwall/btree"
)

type Checkpointer struct {
	savedCheckpoint        atomic.Uint32
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

func (cp *Checkpointer) LatestSavedCheckpoint() uint32 {
	return cp.savedCheckpoint.Load()
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

func (cp *Checkpointer) Checkpoint(writer *ChangesetWriter, root *NodePointer, checkpoint, version uint32, nodeIdsAssigned chan struct{}, seal bool) error {
	if nodeIdsAssigned == nil {
		return fmt.Errorf("nodeIdsAssigned channel cannot be nil when checkpointing, that means we haven't assigned IDs yet")
	}

	select {
	case err := <-cp.doneChan:
		return err
	default:
	}
	cp.reqChan <- checkpointReq{
		writer:          writer,
		root:            root,
		checkpoint:      checkpoint,
		version:         version,
		nodeIdsAssigned: nodeIdsAssigned,
		seal:            seal,
	}
	return nil
}

func (cp *Checkpointer) proc() error {
	var curWriter *ChangesetWriter
	for req := range cp.reqChan {
		_, span := tracer.Start(context.Background(), "SaveCheckpoint")

		// wait for node IDs assignment to complete
		<-req.nodeIdsAssigned
		span.AddEvent("node IDs assigned, proceeding with checkpoint")

		checkpoint := req.checkpoint
		if err := req.writer.SaveCheckpoint(checkpoint, req.version, req.root); err != nil {
			return err
		}
		if err := req.writer.CreateReader(); err != nil {
			return err
		}
		if req.seal {
			if err := req.writer.Seal(); err != nil {
				return err
			}
		}

		// if we have a new writer, update the changeset map
		// we only need to store the changeset ONCE per writer for the FIRST checkpoint it writes
		if req.writer != curWriter { // compare pointers
			curWriter = req.writer
			cp.changesetLock.Lock()
			cp.changesetsByCheckpoint.Set(checkpoint, curWriter.Changeset())
			cp.changesetLock.Unlock()
		}
		cp.savedCheckpoint.Store(checkpoint)
		// notify evictor
		if cp.evictor != nil {
			cp.evictor.Evict(req.root, checkpoint)
		}

		span.End()
	}
	return nil
}

// LatestCheckpointRoot returns the root node pointer and version of the latest saved checkpoint.
// If there are no saved checkpoints, (nil, 0, nil) is returned.
func (cp *Checkpointer) LatestCheckpointRoot() (root *NodePointer, version uint32, err error) {
	cp.changesetLock.RLock()
	defer cp.changesetLock.RUnlock()
	cp.changesetsByCheckpoint.Descend(cp.LatestSavedCheckpoint(), func(checkpoint uint32, cs *Changeset) bool {
		rdr, pin := cs.TryPinReader()
		defer pin.Unpin()
		if rdr == nil {
			err = fmt.Errorf("changeset reader is not available for latest checkpoint %d", checkpoint)
			return false
		}
		root, version = rdr.LatestCheckpointRoot()
		if version != 0 {
			// found a valid checkpoint root
			return false
		}
		// continue searching for an earlier checkpoint with a root
		return true
	})
	return
}

func (cp *Checkpointer) Close() error {
	close(cp.reqChan)
	return <-cp.doneChan
}

type checkpointReq struct {
	writer          *ChangesetWriter
	root            *NodePointer
	checkpoint      uint32
	version         uint32
	seal            bool
	nodeIdsAssigned chan struct{}
}
