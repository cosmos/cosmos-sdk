package internal

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OrphanProcessor struct {
	checkpointer *Checkpointer
	newOrphans   chan *MutationContext
	mtx          sync.Mutex
	errChan      chan error
}

func newOrphanProc(checkpointer *Checkpointer) *OrphanProcessor {
	return &OrphanProcessor{
		checkpointer: checkpointer,
		newOrphans:   make(chan *MutationContext, 32),
		errChan:      make(chan error, 1),
	}
}

func (op *OrphanProcessor) Start() {
	go func() {
		err := op.proc()
		op.errChan <- err
	}()
}

func (op *OrphanProcessor) AddOrphans(mutationCtx *MutationContext) error {
	select {
	case err := <-op.errChan:
		return err
	case op.newOrphans <- mutationCtx:
		return nil
	}
}

func (op *OrphanProcessor) Lock() {
	op.mtx.Lock()
}

func (op *OrphanProcessor) Unlock() {
	op.mtx.Unlock()
}

func (op *OrphanProcessor) proc() error {
	for mutationCtx := range op.newOrphans {
		err := op.procOne(mutationCtx)
		if err != nil {
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
	// sort orphans so that orphan logs are compact and in a deterministic order (sorted by checkpoint, then by flag index)
	slices.SortFunc(orphans, func(a, b *NodePointer) int {
		idA := a.id
		idB := b.id
		c := cmp.Compare(idA.Checkpoint(), idB.Checkpoint())
		if c != 0 {
			return c
		}
		return cmp.Compare(idA.flagIndex, idB.flagIndex)
	})
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
		err := curChangeset.OrphanWriter().WriteOrphan(mutationCtx.version, orphan.id)
		if err != nil {
			return fmt.Errorf("orphan writer failed to write orphan node %s for version %d: %w", orphan.id, mutationCtx.version, err)
		}
		// evict the orphan from memory
		orphan.Mem.Swap(nil) // TODO track memory freed by clearing orphan's in TreeStore.rootMemUsage
	}
	return nil
}

func (op *OrphanProcessor) Close() error {
	close(op.newOrphans)
	return <-op.errChan
}
