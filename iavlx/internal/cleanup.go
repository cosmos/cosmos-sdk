package internal

import (
	"context"
	"slices"
	"sync"
	"time"

	"cosmossdk.io/log/v2"
)

// cleanupProc handles deferred cleanup of changeset resources in a background goroutine.
//
// There are two kinds of cleanup, processed in order:
//
//  1. Disposals (ChangesetReaderRef): when a ChangesetReader is replaced by a newer one
//     (after a checkpoint write or compaction), the old reader is "evicted" and queued here.
//     The cleanup proc periodically tries to dispose it — closing its mmaps once the refcount
//     drops to zero (all pinned readers have unpinned).
//
//  2. Deletions (Changeset): after compaction, the original changeset is queued for deletion.
//     Deletion can only proceed after ALL of its readers have been disposed (activeReaderCount
//     drops to zero). Once that happens, the changeset's directory is deleted from disk.
//
// Disposals are processed before deletions because a changeset can't be deleted until its
// readers are disposed. The sweep runs every second and retries pending items until they succeed.
type cleanupProc struct {
	logger           log.Logger
	mtx              sync.Mutex
	newDisposals     []*ChangesetReaderRef
	newDeletions     []*Changeset
	pendingDisposals []*ChangesetReaderRef
	pendingDeletions []*Changeset
	done             chan struct{}
	cancel           context.CancelFunc
}

func newCleanupProc(logger log.Logger) *cleanupProc {
	return &cleanupProc{
		logger: logger,
		done:   make(chan struct{}),
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
			cp.logger.ErrorContext(ctx, "failed to dispose changeset reader ref", "error", err)
		}
		return disposed
	})

	// we delete after we dispose because disposal frees up resources needed for deletion
	cp.pendingDeletions = slices.DeleteFunc(cp.pendingDeletions, func(cs *Changeset) bool {
		deleted, err := cs.TryDelete(ctx)
		if err != nil {
			cp.logger.ErrorContext(ctx, "failed to delete changeset", "error", err)
		}
		return deleted
	})
}

func (cp *cleanupProc) Close() error {
	cp.cancel()
	<-cp.done
	return nil
}
