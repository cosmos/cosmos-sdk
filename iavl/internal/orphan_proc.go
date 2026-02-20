package internal

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OrphanProcessor struct {
	checkpointer *Checkpointer

	// queueMu protects the queue slice; AddOrphans appends under this lock
	// so it never blocks on procOne/compaction.
	queueMu sync.Mutex
	queue   []*MutationContext
	notify  chan struct{} // capacity-1 signal that new work is available

	// mtx is held during procOne and by compaction (Lock/Unlock) to
	// prevent orphan writes from interfering with changeset rewriting.
	mtx     sync.Mutex
	errChan chan error
}

func newOrphanProc(checkpointer *Checkpointer) *OrphanProcessor {
	return &OrphanProcessor{
		checkpointer: checkpointer,
		notify:       make(chan struct{}, 1),
		errChan:      make(chan error, 1),
	}
}

func (op *OrphanProcessor) Start() {
	go func() {
		err := op.proc()
		op.errChan <- err
	}()
}

// AddOrphans enqueues orphan work without blocking on the orphan processor's
// processing lock. This prevents the commit path from stalling when compaction
// holds the processor lock.
func (op *OrphanProcessor) AddOrphans(mutationCtx *MutationContext) error {
	select {
	case err := <-op.errChan:
		return err
	default:
	}
	op.queueMu.Lock()
	op.queue = append(op.queue, mutationCtx)
	op.queueMu.Unlock()
	// non-blocking signal; if a signal is already pending, the processor
	// will drain the full queue when it wakes up
	select {
	case op.notify <- struct{}{}:
	default:
	}
	return nil
}

func (op *OrphanProcessor) Lock() {
	op.mtx.Lock()
}

func (op *OrphanProcessor) Unlock() {
	op.mtx.Unlock()
}

func (op *OrphanProcessor) proc() error {
	for range op.notify {
		for {
			op.queueMu.Lock()
			if len(op.queue) == 0 {
				op.queueMu.Unlock()
				break
			}
			queue := op.queue
			op.queue = nil
			op.queueMu.Unlock()

			for _, item := range queue {
				if err := op.procOne(item); err != nil {
					return err
				}
			}
		}
	}
	// drain any items remaining after Close signaled
	op.queueMu.Lock()
	remaining := op.queue
	op.queue = nil
	op.queueMu.Unlock()
	for _, item := range remaining {
		if err := op.procOne(item); err != nil {
			return err
		}
	}
	return nil
}

func (op *OrphanProcessor) procOne(mutationCtx *MutationContext) error {
	_, span := tracer.Start(context.Background(),
		"OrphanProcessor.procOne",
		trace.WithAttributes(
			attribute.Int64("version", int64(mutationCtx.version)),
			attribute.Int("numOrphans", len(mutationCtx.orphans)),
		),
	)
	defer span.End()
	orphans := mutationCtx.orphans

	// acquire mutex to ensure we don't interfere with any compaction rewriting
	op.mtx.Lock()
	defer op.mtx.Unlock()
	var curCheckpoint uint32
	var curChangeset *Changeset
	for _, orphan := range orphans {
		checkpoint := orphan.id.Checkpoint()
		if checkpoint != curCheckpoint || curChangeset == nil {
			curChangeset = op.checkpointer.ChangesetByCheckpoint(checkpoint)
			if curChangeset == nil {
				return fmt.Errorf("orphan writer failed to find changeset for checkpoint %d", checkpoint)
			}
			curCheckpoint = checkpoint
		}
		err := curChangeset.OrphanWriter().Append(&OrphanEntry{OrphanedVersion: mutationCtx.version, NodeID: orphan.id})
		if err != nil {
			return fmt.Errorf("orphan writer failed to write orphan node %s for version %d: %w", orphan.id, mutationCtx.version, err)
		}
		// evict the orphan from memory
		orphan.Mem.Swap(nil) // TODO track memory freed by clearing orphan's in TreeStore.rootMemUsage
	}
	return nil
}

func (op *OrphanProcessor) Close() error {
	close(op.notify)
	return <-op.errChan
}
