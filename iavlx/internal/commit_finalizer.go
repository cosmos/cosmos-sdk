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

// CommitFinalizer coordinates the multi-tree commit process.
//
// It is created by CommitBranch.StartCommit, which kicks off per-tree commits in parallel.
// CommitFinalizer then runs its own commit() goroutine that:
//  1. Waits for all per-tree hashes to be ready (prepareCommit)
//  2. Computes the combined multi-tree hash
//  3. Writes commit info to disk (store names + hashes)
//  4. Waits for a finalize or rollback signal from the caller
//  5. Either finalizes all per-tree commits or rolls them all back
//
// The caller interacts with this type through three main methods:
//   - StartFinalize / Finalize: signal that the commit should proceed and wait for completion
//   - Rollback: cancel the commit and undo all per-tree work
//   - SignalFinalize: signal finalization without blocking (used to unblock the coordinator
//     while the caller does other work)
//
// Context hierarchy: the CommitFinalizer has its own cancel context which is a *sibling* of the
// per-tree cancel contexts (they all derive from the same parent tracing span context).
// Canceling the CommitFinalizer's context does not directly cancel per-tree commits.
// Instead, the commit() goroutine detects the cancellation and explicitly rolls back each tree.
type CommitFinalizer struct {
	*CommitMultiTree
	ctx    context.Context
	cancel context.CancelFunc
	// cacheMs is the MultiTree that holds the pending writes from block execution.
	cacheMs *MultiTree
	// finalizers holds one commitTreeFinalizer per IAVL store, each running its own background commit.
	finalizers []*commitTreeFinalizer
	// workingCommitInfo accumulates per-tree hashes as they become ready.
	workingCommitInfo *storetypes.CommitInfo
	// workingCommitId is the combined multi-tree commit hash, set once all per-tree hashes are in.
	workingCommitId storetypes.CommitID
	// done is closed when the commit() goroutine has finished (either successfully or after rollback).
	done chan struct{}
	// hashReady is closed once the combined multi-tree hash has been computed.
	// This allows callers to get the hash before the full commit is finalized (e.g. to return it to CometBFT).
	hashReady chan struct{}
	// finalizeOnce ensures the finalizeOrRollback channel is closed exactly once.
	finalizeOnce sync.Once
	// finalizeOrRollback is closed to signal that the caller has decided to either finalize or rollback.
	// The commit() goroutine blocks on this channel after computing hashes, waiting for the caller's decision.
	finalizeOrRollback chan struct{}
	// err stores any error from the commit() goroutine. Checked after done is closed.
	err atomic.Value
}

// commit is the main coordinator goroutine. It runs in the background and orchestrates
// the full lifecycle of a multi-tree commit: prepare → finalize or rollback.
func (db *CommitFinalizer) commit(ctx context.Context, span trace.Span) error {
	// We pass the span from StartCommit into here and finish it here so that all sub-tree commits
	// are nested under this span.
	defer span.End()

	// Only one commit can be in progress at a time across the entire CommitMultiTree.
	db.commitMutex.Lock()
	defer db.commitMutex.Unlock()

	// prepareCommit does the heavy lifting:
	// - waits for all per-tree hashes
	// - computes the combined hash
	// - writes commit info to disk
	// - waits for the finalize/rollback signal from the caller
	// If anything fails (including the caller requesting a rollback via context cancellation),
	// we need to roll back every per-tree commit.
	if err := db.prepareCommit(ctx); err != nil {
		// Ensure our own context is canceled so any in-flight coordination (writeCommitInfo, etc.) stops.
		db.startRollback()

		// Roll back every per-tree commit in parallel.
		// We use a plain WaitGroup (not errGroup) because we want to attempt ALL rollbacks
		// even if some fail — we don't want one failed rollback to prevent others from running.
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

	// If we get here, prepareCommit succeeded, which means:
	// - All per-tree WAL writes have been fsynced to disk
	// - The commit info file has been atomically renamed and its parent directory fsynced
	// - The caller has signaled finalization
	//
	// At this point, the commit is DURABLE — even if we crash right now, recovery will replay
	// the per-tree WALs to reconstruct the tree state (see ReplayWALForStartup in wal_replay.go).
	//
	// What's left for per-tree finalization is NOT about durability, it's about updating in-memory
	// state so the running process can see the new version:
	// - Atomically swap the in-memory root pointer to the new tree root (SaveRoot)
	// - Kick off background checkpoint writing (an optimization that makes future startup faster
	//   by avoiding full WAL replay, but NOT required for correctness)
	// - Queue orphan nodes for later cleanup
	// - Update the lastCommitId field
	var errGroup errgroup.Group
	// Finalize IAVL stores — unblocks each per-tree commit() goroutine so it can run SaveRoot.
	for _, finalizer := range db.finalizers {
		errGroup.Go(func() error {
			_, err := finalizer.Finalize()
			return err
		})
	}
	// Commit non-IAVL stores (mem, transient, etc.) — these are simpler stores that just flush their caches.
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
	// Wait for all stores to finalize.
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("finalizing commit failed: %w", err)
	}

	// Update the CommitMultiTree's committed state so that subsequent reads see the new version.
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

// writeCommitInfo writes the commit info file in two phases to overlap IO with hash computation:
//
// Phase 1 (writeCommitInfoHeader): Write store names → fsync → wait for all per-tree WALs → atomic rename + dir fsync.
// This is the durability boundary. Once headerDone returns nil, the commit is crash-recoverable.
// The per-tree WALs contain all the data needed to reconstruct the trees; the commit info file
// records which version was committed so recovery knows which WAL entries to replay.
//
// Phase 2 (footer): Append per-tree hashes to the already-renamed file. These are NOT fsynced
// because they're an optimization (avoids recomputing hashes on load), not required for correctness.
// If we crash before the footer is written, recovery recomputes hashes from the replayed trees.
func (db *CommitFinalizer) writeCommitInfo(headerDone chan error) {

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

// writeCommitInfoHeader implements the durability protocol:
//  1. Serialize the commit info header (store names) to a buffer
//  2. Wait for the finalize signal (so we don't write to disk for a commit that gets rolled back)
//  3. Write header to a .pending file and fsync it
//  4. Wait for ALL per-tree WAL writes to complete (this is the critical ordering constraint)
//  5. Atomic rename .pending → final (this is the moment the commit becomes crash-recoverable)
//  6. Fsync the parent directory to make the rename durable
//
// Returns the open file handle so the caller (writeCommitInfo) can append hashes later.
func (db *CommitFinalizer) writeCommitInfoHeader() (*os.File, error) {
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

	// Fsync the .pending file BEFORE renaming it to the final path.
	// This follows the standard "write temp, fsync, rename" pattern used by most databases
	// (SQLite, LevelDB/RocksDB, etcd, Postgres WAL segments, etc.) for crash-safe file updates.
	// Rename is atomic (the file either appears at the new path or doesn't), but it does NOT
	// guarantee the file's contents are on disk. Without this fsync, a crash right after the
	// rename could leave a zero-length or corrupt file at the final path — recovery would see
	// the commit info file exists (the rename was durable) but find garbage or nothing inside it.
	// By fsyncing first, we guarantee that if the rename survives a crash, so do the contents.
	err = file.Sync()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to sync commit info file for version %d: %w", stagedVersion, err)
	}

	// Wait for ALL per-tree WAL writes to complete before renaming.
	// This is the critical ordering constraint: the rename is what makes the commit "exist" from
	// recovery's perspective, so all per-tree WALs must be fsynced BEFORE we rename.
	// After the rename + dir fsync below, a crash at any point is recoverable:
	// recovery sees the commit info file, then replays each tree's WAL to reconstruct the committed state.
	// If we crash BEFORE the rename, the per-tree WAL entries for this version are simply
	// truncated away during recovery (see ReplayWALForStartup auto-repair).
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

	// Atomic rename: this is the moment the commit becomes visible to crash recovery.
	err = os.Rename(pendingPath, commitInfoPath)
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to rename commit info file for version %d: %w", stagedVersion, err)
	}

	// Fsync the parent directory to ensure the rename is durable on disk (not just in the filesystem journal).
	// This runs while per-tree hash computation may still be in progress,
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

// prepareCommit does the "prepare" phase of the two-phase commit:
//  1. Kick off commit info writing in the background (so fsync can overlap with hash computation)
//  2. Wait for every per-tree commit to produce its root hash
//  3. Compute the combined multi-tree hash and signal hashReady
//  4. Wait for the caller to signal finalize or rollback
//  5. Wait for commit info to be durably written before returning
//
// If this method returns nil, we are past the point of no return — the caller MUST finalize.
// If it returns an error, the caller MUST roll back all per-tree commits.
func (db *CommitFinalizer) prepareCommit(ctx context.Context) error {
	// Start writing commit info (store names, then hashes) to disk in a background goroutine.
	// This overlaps the expensive fsync with hash computation so we don't pay for both sequentially.
	// See writeCommitInfo for details on the two-phase write (header first, then footer with hashes).
	commitInfoSynced := make(chan error, 1)
	go func() {
		db.writeCommitInfo(commitInfoSynced)
	}()

	// Wait for each per-tree commit to produce its root hash.
	// Each per-tree commitTreeFinalizer computes its hash independently in its own goroutine,
	// and we collect them all here in parallel.
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

	// All per-tree hashes are in. Compute the combined multi-tree hash.
	db.workingCommitId = storetypes.CommitID{
		Version: db.stagedVersion(),
		Hash:    db.workingCommitInfo.Hash(),
	}
	// Signal that the hash is ready — callers waiting on WaitForHash / StartFinalize can now proceed.
	// This is important because it allows FinalizeBlock to return the app hash to CometBFT
	// before the full commit is done.
	close(db.hashReady)

	// Now we wait for the caller's decision: finalize or rollback.
	// The caller signals by closing the finalizeOrRollback channel (via SignalFinalize or Rollback).
	// If the context is canceled (e.g. external Rollback call), we also unblock.
	select {
	case <-db.finalizeOrRollback:
	case <-ctx.Done():
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	// Wait for commit info to be durably written to disk before we let stores finalize.
	// writeCommitInfoHeader waits for ALL per-tree WAL writes to complete, then atomically renames
	// the commit info file and fsyncs the directory. This is the durability boundary:
	// once this returns, the commit is recoverable from a crash via WAL replay.
	//
	// We must wait for this before letting per-tree finalization proceed because finalization
	// can trigger background checkpointing. If a checkpoint were written before the commit info
	// file exists, a crash could leave us in an inconsistent state where a tree's checkpoint
	// has advanced but we have no commit info record for that version.
	if err := <-commitInfoSynced; err != nil {
		return fmt.Errorf("writing commit info failed: %w", err)
	}

	// We are past the point of no return — the commit info is durable and the caller has signaled finalize.
	// We intentionally do NOT return ctx.Err() here, because even if the context was canceled
	// between the select and now, we've already committed to finalizing.
	return nil
}

// StartFinalize signals finalization and blocks until the combined multi-tree hash is available.
// Returns as soon as the hash is ready — the full commit (WAL fsync, commit info write, per-tree
// finalization) may still be in progress. This is the method used by FinalizeBlock to return
// the app hash to CometBFT as early as possible.
func (db *CommitFinalizer) StartFinalize() (storetypes.CommitID, error) {
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

// SignalFinalize unblocks the commit coordinator without waiting for any result.
// This closes the finalizeOrRollback channel, which unblocks prepareCommit's wait
// and allows the commit to proceed to durability and per-tree finalization.
func (db *CommitFinalizer) SignalFinalize() error {
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
	return nil
}

// Finalize signals finalization and blocks until the ENTIRE commit is complete —
// all per-tree WALs fsynced, commit info durable, in-memory roots swapped, checkpoints kicked off.
// Use this when you need to be certain the commit is fully done before proceeding.
func (db *CommitFinalizer) Finalize() (storetypes.CommitID, error) {
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

// Rollback cancels the commit and waits for all cleanup to complete.
// Per-tree WAL rollback is just a truncation back to the previous offset (append-only WALs make this cheap).
// The commit info file is never renamed (or if we haven't gotten that far, the .pending file is cleaned up).
func (db *CommitFinalizer) Rollback() error {
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

// startRollback triggers rollback by canceling the CommitFinalizer's context and unblocking
// any goroutines waiting on the finalizeOrRollback channel.
// Note: this does NOT directly cancel per-tree contexts (they are siblings, not children).
// The commit() goroutine detects the cancellation, then explicitly calls Rollback() on each
// per-tree finalizer, which cancels their individual contexts and truncates their WALs.
func (db *CommitFinalizer) startRollback() {
	db.cancel()
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
}
