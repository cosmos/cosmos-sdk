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

type CommitBranch struct {
	// MultiTree is the cache layer with staged writes
	*MultiTree
	db *CommitMultiTree
}

func (cb *CommitBranch) StartCommit(ctx context.Context, header cmtproto.Header) (*MultiTreeFinalizer, error) {
	db := cb.db
	ctx, span := tracer.Start(ctx, "CommitMultiTree.commit",
		trace.WithAttributes(
			attribute.Int64("version", db.stagedVersion()),
		),
	)

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
	for i, si := range db.iavlStores {
		commitStore := si.store.(*CommitTree)
		cachedStore := multiTree.GetCacheWrapIfExists(si.key)
		var updates iter.Seq[KVUpdate]
		var updateCount int
		if cachedStore != nil {
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
	ctx, cancel := context.WithCancel(ctx)
	finalizer := &MultiTreeFinalizer{
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
