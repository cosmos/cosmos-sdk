package internal

import (
	"context"
	"fmt"
	"iter"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/iavlx/internal/cachekv"
)

// CommitBranch is the bridge between a MultiTree (an in-memory cache of pending writes) and a
// CommitMultiTree (the persistent, committed state of all IAVL trees).
//
// The idea is that during block execution, all store mutations happen against the MultiTree cache.
// When the block is done, CommitBranch.StartCommit kicks off background commits for every individual
// IAVL tree in parallel, returning a CommitFinalizer that the caller can use to either finalize
// (make the commit permanent) or rollback (discard everything) — this is the "optimistic commit" pattern.
//
// The optimistic part: we start doing expensive commit work (tree mutations, WAL writes, hashing)
// before we know for sure that this block will be accepted. If consensus rejects the block,
// we can roll everything back cheaply. If it accepts, we've already done most of the work.
type CommitBranch struct {
	// MultiTree holds the cached writes from block execution that haven't been committed yet.
	*MultiTree
	// db is the underlying persistent multi-tree that we're committing to.
	db *CommitMultiTree
}

// StartCommit begins the optimistic commit process for all IAVL trees in parallel.
//
// Here's the high-level flow:
//  1. Sanity-check that our cache is based on the latest committed version (no stale reads).
//  2. For each IAVL store, extract the pending writes from the cache and kick off a background
//     commit via CommitTree.startCommit. Each tree commit runs independently and in parallel.
//  3. Spin up a top-level goroutine that coordinates waiting for all per-tree commits,
//     computing the combined hash, and handling finalize/rollback signals.
//  4. Return a CommitFinalizer that the caller uses to either finalize or rollback.
//
// At this point, all the expensive work (tree mutations, WAL writes, hashing) is happening in the
// background. The caller can do other work (like returning a hash to CometBFT) while commits proceed.
func (cb *CommitBranch) StartCommit(ctx context.Context, header cmtproto.Header) (*CommitFinalizer, error) {
	db := cb.db
	ctx, span := tracer.Start(ctx, "CommitMultiTree.commit",
		trace.WithAttributes(
			attribute.Int64("version", db.stagedVersion()),
		),
	)

	// Guard against committing stale state: the cache must have been created from the latest version.
	latestVersion := db.LatestVersion()
	multiTree := cb.MultiTree
	if multiTree.LatestVersion() != latestVersion {
		return nil, fmt.Errorf("store version mismatch: expected %d, got %d", latestVersion, multiTree.LatestVersion())
	}

	numIavlStores := len(db.iavlStores)
	storeInfos := make([]storetypes.StoreInfo, numIavlStores)
	finalizers := make([]*commitTreeFinalizer, numIavlStores)
	commitInfo := &storetypes.CommitInfo{
		StoreInfos: storeInfos,
		Timestamp:  header.Time,
		Version:    db.stagedVersion(),
	}

	// For each IAVL store, pull pending writes from the cache and start committing in the background.
	// Each CommitTree.startCommit returns a commitTreeFinalizer which lets us wait for the hash,
	// signal finalization, or trigger a rollback — all independently per tree.
	for i, si := range db.iavlStores {
		commitStore := si.store.(*CommitTree)
		cachedStore := multiTree.GetCacheWrapIfExists(si.key)
		var updates iter.Seq[KVUpdate]
		var updateCount int
		if cachedStore != nil {
			// Only stores that were actually touched during block execution will have a cache entry.
			// Untouched stores get nil updates, meaning an empty commit (just a version bump).
			cacheKv, ok := cachedStore.(*cachekv.Store)
			if !ok {
				return nil, fmt.Errorf("expected %T, got %T", &cachekv.Store{}, cachedStore)
			}
			updates, updateCount = cacheKv.Updates()
		}
		finalizer := commitStore.startCommit(ctx, updates, updateCount)
		finalizers[i] = finalizer
		storeInfos[i].Name = si.key.Name()
	}

	// Create a cancellable context for the CommitFinalizer's own coordination logic
	// (prepareCommit, writeCommitInfo, etc.).
	// Note: this is a sibling of the per-tree cancel contexts, NOT their parent.
	// Canceling this context does NOT directly cancel per-tree commits.
	// Instead, when this context is canceled (via Rollback), the commit() goroutine notices,
	// and explicitly calls Rollback() on each per-tree finalizer, which cancels their individual contexts.
	ctx, cancel := context.WithCancel(ctx)
	finalizer := &CommitFinalizer{
		CommitMultiTree:    db,
		cacheMs:            multiTree,
		ctx:                ctx,
		cancel:             cancel,
		finalizers:         finalizers,
		workingCommitInfo:  commitInfo,
		done:               make(chan struct{}),
		hashReady:          make(chan struct{}),
		finalizeOrRollback: make(chan struct{}),
	}

	// The commit coordinator goroutine: waits for all per-tree commits to produce hashes,
	// computes the combined multi-tree hash, then waits for the finalize/rollback signal.
	// See CommitFinalizer.commit for the details.
	go func() {
		// Prevent context leak: WithCancel registers a child in the parent context's tree,
		// and that registration is only cleaned up when cancel() is called.
		// Safe here because all ctx.Err() checks are inside commit(), which has already
		// returned by the time this defer fires. On rollback, cancel() is called first;
		// calling it again here is a no-op.
		defer cancel()
		err := finalizer.commit(ctx, span)
		if err != nil {
			finalizer.err.Store(err)
		}
		close(finalizer.done)
		db.compactIfNeeded() // start background compaction when needed
	}()
	return finalizer, nil
}

// commitBranchCacheCompatWrapper is a wrapper around a CommitBranch
// which allows the CacheMultiStore.Write() method to be used as a way to
// start a background commit in a way that is compatible with store v1 and v2.
type commitBranchCacheCompatWrapper struct {
	*CommitBranch
}

// Write starts a background commit by calling StartCommit and then immediately signaling finalization.
func (wrapper *commitBranchCacheCompatWrapper) Write() {
	wrapper.db.compatFinalizerMu.Lock()
	defer wrapper.db.compatFinalizerMu.Unlock()

	if wrapper.db.compatFinalizer != nil {
		panic("Write has already been called on this CacheMultiStore, it cannot be called twice when using iavlx")
	}
	finalizer, err := wrapper.StartCommit(context.Background(), cmtproto.Header{})
	if err != nil {
		panic(err)
	}
	// We signal finalization because in the store/v1 CommitMultiStore calling Write is a point of no return anyway,
	// so this gives us a bit of a speed up.
	err = finalizer.SignalFinalize()
	if err != nil {
		panic(err)
	}
	wrapper.db.compatFinalizer = finalizer
}

var _ storetypes.CacheMultiStore = (*commitBranchCacheCompatWrapper)(nil)
