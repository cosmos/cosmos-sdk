package internal

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/tidwall/btree"
)

type Checkpointer struct {
	savedLayer          atomic.Uint32
	reqChan             chan checkpointReq
	doneChan            chan error
	changesetLock       sync.RWMutex
	changesetsByLayer   *btree.Map[uint32, *Changeset]
	changesetsByVersion *btree.Map[uint32, *Changeset]
	evictor             Evictor
}

func NewCheckpointer(evictor Evictor) *Checkpointer {
	cp := &Checkpointer{
		reqChan:             make(chan checkpointReq, 16),
		doneChan:            make(chan error, 1),
		changesetsByLayer:   &btree.Map[uint32, *Changeset]{},
		changesetsByVersion: &btree.Map[uint32, *Changeset]{},
		evictor:             evictor,
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

func (cp *Checkpointer) ChangesetByLayer(layer uint32) *Changeset {
	cp.changesetLock.RLock()
	defer cp.changesetLock.RUnlock()

	var res *Changeset
	// Find the changeset with the highest start version <= the requested layer
	cp.changesetsByLayer.Descend(layer, func(key uint32, cs *Changeset) bool {
		res = cs
		return false // Take the first (highest) entry <= layer
	})
	return res
}

func (cp *Checkpointer) ChangesetByVersion(version uint32) *Changeset {
	cp.changesetLock.RLock()
	defer cp.changesetLock.RUnlock()

	var res *Changeset
	// Find the changeset with the highest start version <= the requested version
	cp.changesetsByVersion.Descend(version, func(key uint32, cs *Changeset) bool {
		res = cs
		return false // Take the first (highest) entry <= version
	})
	return res
}

func (cp *Checkpointer) Checkpoint(writer *ChangesetWriter, root *NodePointer, version uint32, seal bool) {
	cp.reqChan <- checkpointReq{
		writer:  writer,
		root:    root,
		version: version,
		seal:    seal,
	}
}

func (cp *Checkpointer) proc() error {
	var curWriter *ChangesetWriter
	for req := range cp.reqChan {
		_, span := Tracer.Start(context.Background(), "SaveCheckpoint")

		layer := cp.savedLayer.Load() + 1
		if err := req.writer.SaveLayer(layer, req.root); err != nil {
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
		// if we have a new writer, update the changeset maps
		// we only need to store the changeset ONCE per writer for the FIRST layer and version it writes
		if req.writer != curWriter { // compare pointers
			curWriter = req.writer
			cp.changesetLock.Lock()
			cp.changesetsByLayer.Set(layer, curWriter.Changeset())
			cp.changesetsByVersion.Set(req.version, curWriter.Changeset())
			cp.changesetLock.Unlock()
		}
		cp.savedLayer.Store(layer)
		// notify evictor
		if cp.evictor != nil {
			cp.evictor.Evict(req.root, layer)
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
