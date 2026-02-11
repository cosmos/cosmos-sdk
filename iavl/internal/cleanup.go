package internal

import (
	"context"
	"slices"
	"sync"
	"time"
)

type cleanupProc struct {
	mtx              sync.Mutex
	newDisposals     []*ChangesetReaderRef
	newDeletions     []*Changeset
	pendingDisposals []*ChangesetReaderRef
	pendingDeletions []*Changeset
	done             chan struct{}
	cancel           context.CancelFunc
}

func newCleanupProc() *cleanupProc {
	return &cleanupProc{
		done: make(chan struct{}),
	}
}

func (cp *cleanupProc) AddDisposal(ref *ChangesetReaderRef) {
	cp.mtx.Lock()
	defer cp.mtx.Unlock()
	cp.newDisposals = append(cp.newDisposals, ref)
}

func (cp *cleanupProc) AddDeletion(changeset *Changeset) {
	cp.mtx.Lock()
	defer cp.mtx.Unlock()
	cp.newDeletions = append(cp.newDeletions, changeset)
}

func (cp *cleanupProc) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	cp.cancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				// wait a tiny bit, then do a final sweep
				time.Sleep(100 * time.Millisecond)
				cp.sweep(ctx)
				close(cp.done)
				return
			case <-time.After(1 * time.Second):
			}

			cp.sweep(ctx)
		}
	}()
}

func (cp *cleanupProc) sweep(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "cleanupProc.sweep")
	defer span.End()

	// drain incoming
	cp.mtx.Lock()
	disposals := cp.newDisposals
	deletions := cp.newDeletions
	cp.newDisposals = nil
	cp.newDeletions = nil
	cp.mtx.Unlock()

	cp.pendingDisposals = append(cp.pendingDisposals, disposals...)
	cp.pendingDeletions = append(cp.pendingDeletions, deletions...)

	cp.pendingDisposals = slices.DeleteFunc(cp.pendingDisposals, func(ref *ChangesetReaderRef) bool {
		disposed, err := ref.TryDispose()
		if err != nil {
			logger.ErrorContext(ctx, "failed to dispose changeset reader ref", "error", err)
		}
		return disposed
	})

	// we delete after we dispose because disposal frees up resources needed for deletion
	cp.pendingDeletions = slices.DeleteFunc(cp.pendingDeletions, func(cs *Changeset) bool {
		deleted, err := cs.TryDelete(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "failed to delete changeset", "error", err)
		}
		return deleted
	})
}

func (cp *cleanupProc) Close() error {
	cp.cancel()
	<-cp.done
	return nil
}
