package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

type MultiTreeFinalizer struct {
	*CommitMultiTree
	ctx                context.Context
	cancel             context.CancelFunc
	cacheMs            *MultiTree
	finalizers         []*commitTreeFinalizer
	workingCommitInfo  *storetypes.CommitInfo
	workingCommitId    storetypes.CommitID
	done               chan struct{}
	hashReady          chan struct{}
	finalizeOnce       sync.Once
	finalizeOrRollback chan struct{}
	err                atomic.Value
}

func (db *MultiTreeFinalizer) commit(ctx context.Context, span trace.Span) error {
	// we pass the span from StartCommit into here and finish it here so that all sub-tree commits
	// are nested under this span
	defer span.End()

	db.commitMutex.Lock()
	defer db.commitMutex.Unlock()

	if err := db.prepareCommit(ctx); err != nil {
		db.startRollback()

		// do not use an errGroup here since, we want to rollback everything even if some rollbacks fail
		var wg sync.WaitGroup
		errs := make([]error, len(db.finalizers))
		for i, finalizer := range db.finalizers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[i] = finalizer.Rollback()
			}()
		}

		wg.Wait()
		if err := errors.Join(errs...); err != nil {
			return fmt.Errorf("rollback failed: %w", err.(error))
		}

		return fmt.Errorf("%w; cause: %v", rolledbackErr, err)
	}

	var errGroup errgroup.Group
	// finalize IAVL stores
	for _, finalizer := range db.finalizers {
		errGroup.Go(func() error {
			_, err := finalizer.Finalize()
			return err
		})
	}
	// commit non-IAVL stores
	for _, si := range db.otherStores {
		errGroup.Go(func() error {
			cachedStore := db.cacheMs.GetCacheWrapIfExists(si.key)
			if cachedStore == nil {
				return nil
			}
			cachedStore.Write()
			committer, ok := si.store.(storetypes.Committer)
			if !ok {
				return nil
			}
			committer.Commit()
			return nil
		})
	}
	// wait for all stores to finalize
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("finalizing commit failed: %w", err)
	}

	version := db.workingCommitId.Version
	db.commitData.Store(&commitData{
		commitId:   db.workingCommitId,
		commitInfo: db.workingCommitInfo,
	})
	if db.earliestVersion.Load() == 0 {
		db.earliestVersion.Store(version)
	}
	return nil
}

func (db *MultiTreeFinalizer) writeCommitInfo(headerDone chan error) {
	// in order to not block on fsync until AFTER we have computed all hashes, which SHOULD be the slowest operation (WAL writing should complete before that)
	// we write and fsync the first part of the commit info (store names) as soon as we know finalization will happen,
	// and then append hashes at the end once they are ready, without fsyncing again since they aren't needed for durability

	file, err := db.writeCommitInfoHeader()
	headerDone <- err
	close(headerDone)
	if err != nil {
		return
	}

	// Wait for hashes to be ready. The ctx.Done case prevents a goroutine leak:
	// if SignalFinalize() was called before hashes completed and hash computation
	// then fails, hashReady is never closed and this goroutine would block forever.
	// At this point the durable state is already settled (committed or rolled back),
	// so we just clean up and exit.
	select {
	case <-db.hashReady:
	case <-db.ctx.Done():
		_ = file.Close()
		return
	}

	err = writeCommitInfoFooter(file, db.workingCommitInfo)
	if err != nil {
		// at this point we don't error on such errors
		db.logger.Error("failed to write commit info footer with hashes", "error", err)
	}

	err = file.Close()
	if err != nil {
		db.logger.Error("failed to close commit info file after writing hashes", "error", err)
		return
	}
}

func (db *MultiTreeFinalizer) writeCommitInfoHeader() (*os.File, error) {
	var headerBuf bytes.Buffer
	info := db.workingCommitInfo
	err := writeCommitInfoHeader(&headerBuf, info)
	if err != nil {
		return nil, fmt.Errorf("failed to write commit info header to buffer: %w", err)
	}

	// wait for finalization signal
	// TODO this wait can be moved down to right before the rename and that would be more efficient -
	// we would do all of the IO heavy work (specifically fsync) while waiting for the finalization signal.
	// That would mean even less chance that we introduce any latency here. We would just need to make sure
	// the .pending.* file gets cleaned up correctly.
	select {
	case <-db.finalizeOrRollback:
	case <-db.ctx.Done():
	}
	if db.ctx.Err() != nil {
		return nil, db.ctx.Err() // do not write commit info if rolling back
	}

	// write the header to disk
	commitInfoDir := filepath.Join(db.dir, commitInfoSubPath)
	err = os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	stagedVersion := info.Version

	pendingPath := filepath.Join(commitInfoDir, fmt.Sprintf(".pending.%d", stagedVersion))
	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", stagedVersion))
	file, err := os.OpenFile(pendingPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open commit info file for version %d: %w", stagedVersion, err)
	}

	_, err = file.Write(headerBuf.Bytes())
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to write commit info header for version %d: %w", stagedVersion, err)
	}

	// fsync the file to ensure durability of store names
	err = file.Sync()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to sync commit info file for version %d: %w", stagedVersion, err)
	}

	// wait for all trees to complete their WAL writes before renaming so we only commit when all children have committed
	var wg errgroup.Group
	for _, finalizer := range db.finalizers {
		wg.Go(func() error {
			return finalizer.WaitForWAL()
		})
	}
	if err := wg.Wait(); err != nil {
		_ = file.Close()
		_ = os.Remove(pendingPath)
		return nil, fmt.Errorf("failed when waiting for WAL completion: %w", err)
	}

	err = os.Rename(pendingPath, commitInfoPath)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to rename commit info file for version %d: %w", stagedVersion, err)
	}

	// fsync the parent directory to ensure the rename is durable.
	// This runs while per-tree hash computation is still in progress,
	// so it adds no latency to the critical path.
	parentDir, err := os.Open(commitInfoDir)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to open commit info dir for fsync: %w", err)
	}
	if err := parentDir.Sync(); err != nil {
		_ = parentDir.Close()
		_ = file.Close()
		return nil, fmt.Errorf("failed to fsync commit info dir: %w", err)
	}
	_ = parentDir.Close()

	return file, nil
}

func (db *MultiTreeFinalizer) prepareCommit(ctx context.Context) error {
	// start writing commit info in background
	commitInfoSynced := make(chan error, 1)
	go func() {
		db.writeCommitInfo(commitInfoSynced)
	}()

	var hashErrGroup errgroup.Group
	for i, finalizer := range db.finalizers {
		hashErrGroup.Go(func() error {
			hash, err := finalizer.WaitForHash()
			if err != nil {
				return err
			}
			if hash.Version != 0 && hash.Version != int64(db.stagedVersion()) {
				return fmt.Errorf("store %s returned mismatched version in commit ID: expected %d, got %d", db.iavlStores[i].key.Name(), db.stagedVersion(), hash.Version)
			}
			db.workingCommitInfo.StoreInfos[i].CommitId = hash
			return nil
		})
	}
	if err := hashErrGroup.Wait(); err != nil {
		return err
	}

	db.workingCommitId = storetypes.CommitID{
		Version: db.stagedVersion(),
		Hash:    db.workingCommitInfo.Hash(),
	}
	close(db.hashReady)

	select {
	case <-db.finalizeOrRollback:
	case <-ctx.Done():
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	// wait for commit info to be written before we start finalizing stores,
	// otherwise checkpointing may start, and commit is not atomic
	if err := <-commitInfoSynced; err != nil {
		return fmt.Errorf("writing commit info failed: %w", err)
	}

	// we are past the rollback point so we don't return ctx.Err()
	return nil
}

func (db *MultiTreeFinalizer) StartFinalize() (storetypes.CommitID, error) {
	if err := db.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}
	select {
	case <-db.hashReady:
	case <-db.done:
	}
	err := db.err.Load()
	if err != nil {
		return storetypes.CommitID{}, err.(error)
	}
	return db.workingCommitId, nil
}

func (db *MultiTreeFinalizer) SignalFinalize() error {
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
	return nil
}

func (db *MultiTreeFinalizer) Finalize() (storetypes.CommitID, error) {
	if err := db.SignalFinalize(); err != nil {
		return storetypes.CommitID{}, err
	}

	<-db.done
	err := db.err.Load()
	if err != nil {
		return storetypes.CommitID{}, err.(error)
	}
	return db.workingCommitId, nil
}

func (db *MultiTreeFinalizer) Rollback() error {
	db.startRollback()
	<-db.done
	err := db.err.Load()
	if err == nil {
		return fmt.Errorf("rollback failed, commit succeeded")
	}
	if !errors.Is(err.(error), rolledbackErr) {
		return fmt.Errorf("rollback failed: %w", err.(error))
	}
	return nil
}

func (db *MultiTreeFinalizer) startRollback() {
	// we must propagate cancellation to any background operations
	db.cancel()
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
}
