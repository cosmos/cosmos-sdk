package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cosmos/btree"
)

// Checkpointer manages background checkpoint writing and orchestrates the post-commit pipeline.
//
// After a commit finalizes and SaveRoot updates the in-memory root, it sends a request to the
// Checkpointer via the Checkpoint or QueueOrphans methods. The Checkpointer's background
// goroutine (proc) processes these requests sequentially:
//
//  1. Send orphans to the OrphanProcessor (always, even if no checkpoint is being written).
//  2. If a checkpoint is requested: wait for NodeID assignment to complete, then call
//     ChangesetWriter.SaveCheckpoint to persist the tree to disk.
//  3. Create/update the ChangesetReader so the persisted data is queryable.
//  4. If the changeset needs to be sealed (rollover), seal it.
//  5. Register the changeset in the checkpoint lookup map.
//  6. Notify the Evictor to free memory for nodes that are now on disk.
//
// Not every commit produces a checkpoint — the CheckpointInterval option controls how often
// checkpoints are written. Between checkpoints, only orphan processing happens (step 1).
// The WAL provides durability regardless; checkpoints are an optimization to reduce WAL replay
// time on startup.
type Checkpointer struct {
	savedCheckpoint        atomic.Uint32
	reqChan                chan checkpointReq
	doneChan               chan error
	changesetLock          sync.RWMutex
	changesetsByCheckpoint *btree.Map[uint32, *Changeset]
	evictor                Evictor
	orphanProc             *OrphanProcessor
}

func NewCheckpointer(evictor Evictor) *Checkpointer {
	cp := &Checkpointer{
		reqChan:                make(chan checkpointReq, 32),
		doneChan:               make(chan error, 1),
		changesetsByCheckpoint: &btree.Map[uint32, *Changeset]{},
		evictor:                evictor,
	}
	cp.orphanProc = newOrphanProc(cp)
	cp.start()
	cp.orphanProc.Start()
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

func (cp *Checkpointer) QueueOrphans(mutationCtx *MutationContext) error {
	select {
	case err := <-cp.doneChan:
		return err
	default:
	}
	cp.reqChan <- checkpointReq{
		mutationCtx: mutationCtx,
	}
	return nil
}

func (cp *Checkpointer) Checkpoint(mutationCtx *MutationContext, writer *ChangesetWriter, root *NodePointer, checkpoint uint32, nodeIdsAssigned chan struct{}, seal bool) error {
	if nodeIdsAssigned == nil {
		return fmt.Errorf("nodeIdsAssigned channel cannot be nil when checkpointing, that means we haven't assigned IDs yet")
	}

	select {
	case err := <-cp.doneChan:
		return err
	default:
	}
	cp.reqChan <- checkpointReq{
		mutationCtx:     mutationCtx,
		writer:          writer,
		root:            root,
		checkpoint:      checkpoint,
		nodeIdsAssigned: nodeIdsAssigned,
		seal:            seal,
	}
	return nil
}

func (cp *Checkpointer) proc() error {
	var curWriter *ChangesetWriter
	for req := range cp.reqChan {
		// send orphans to orphan processor
		err := cp.orphanProc.AddOrphans(req.mutationCtx)
		if err != nil {
			return fmt.Errorf("failed to add orphans to orphan processor: %w", err)
		}

		// TODO instrument any channel send delays here
		// if we don't have a writer we're not saving any checkpoints, just doing orphan processing, so skip the checkpoint saving logic
		if req.writer == nil {
			continue
		}

		_, span := tracer.Start(context.Background(), "SaveCheckpoint")

		// wait for node IDs assignment to complete
		<-req.nodeIdsAssigned
		span.AddEvent("node IDs assigned, proceeding with checkpoint")

		checkpoint := req.checkpoint
		if err := req.writer.SaveCheckpoint(checkpoint, req.mutationCtx.version, req.root); err != nil {
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
// If there are no saved checkpoints, a zero-value CheckpointRootInfo is returned.
func (cp *Checkpointer) LatestCheckpointRoot() (res CheckpointRootInfo, err error) {
	cp.changesetLock.RLock()
	defer cp.changesetLock.RUnlock()
	cp.changesetsByCheckpoint.Descend(cp.LatestSavedCheckpoint(), func(checkpoint uint32, cs *Changeset) bool {
		rdr, pin := cs.TryPinReader()
		defer pin.Unpin()
		if rdr == nil {
			err = fmt.Errorf("changeset reader is not available for latest checkpoint %d", checkpoint)
			return false
		}
		res = rdr.LatestCheckpointRoot()
		if res.Version != 0 {
			// we found a checkpoint with a valid root, stop iterating and return it
			return false
		}
		// continue earlier checkpoints for a valid root
		return true
	})
	return res, err
}

func (cp *Checkpointer) Close() error {
	close(cp.reqChan)
	err := <-cp.doneChan
	// close orphan proc after checkpointer proc finishes, since the
	// checkpointer proc sends to the orphan proc's channel
	return errors.Join(err, cp.orphanProc.Close())
}

type checkpointReq struct {
	writer          *ChangesetWriter
	mutationCtx     *MutationContext
	root            *NodePointer
	checkpoint      uint32
	seal            bool
	nodeIdsAssigned chan struct{}
}
