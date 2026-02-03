package iavl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/store/mem"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	"cosmossdk.io/store/transient"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/iavl/internal"
	"github.com/cosmos/cosmos-sdk/iavl/internal/cachekv"
)

type CommitMultiTree struct {
	dir        string
	opts       Options
	trees      []storetypes.CacheWrapper   // always ordered by tree name
	treeKeys   []storetypes.StoreKey       // always ordered by tree name
	storeTypes []storetypes.StoreType      // store types by tree index
	treesByKey map[storetypes.StoreKey]int // index of the trees by name

	commitMutex    sync.Mutex
	version        uint64
	lastCommitId   storetypes.CommitID
	lastCommitInfo *storetypes.CommitInfo
}

func (db *CommitMultiTree) StartCommit(ctx context.Context, store storetypes.MultiStore, header cmtproto.Header) (storetypes.CommitFinalizer, error) {
	// TODO add mutex if needed
	multiTree, ok := store.(*internal.MultiTree)
	if !ok {
		return nil, fmt.Errorf("expected MultiTree, got %T", store)
	}

	if multiTree.LatestVersion() != int64(db.version) {
		return nil, fmt.Errorf("store version mismatch: expected %d, got %d", db.version, multiTree.LatestVersion())
	}

	storeInfos := make([]storetypes.StoreInfo, len(db.trees))
	finalizers := make([]storetypes.CommitFinalizer, len(db.trees))
	commitInfo := &storetypes.CommitInfo{
		StoreInfos: storeInfos,
		Timestamp:  header.Time,
	}
	for i, treeKey := range db.treeKeys {
		storeInfos[i].Name = treeKey.Name()
		cachedStore := multiTree.GetCacheWrapIfExists(treeKey)
		tree := db.trees[i]
		switch commitStore := tree.(type) {
		case *CommitTree:
			var updates iter.Seq[KVUpdate]
			if cachedStore != nil {
				cacheKv, ok := cachedStore.(*cachekv.Store)
				if !ok {
					return nil, fmt.Errorf("expected cachekv.Store, got %T", cachedStore)
				}
				updates = cacheKv.Updates()
			}
			finalizers[i] = commitStore.StartCommit(ctx, updates, 0)
		case *mem.Store, *transient.Store, *transient.ObjStore:
			finalizers[i] = &fauxCommitStoreFinalizer{
				committer: commitStore.(storetypes.Committer),
				cacheWrap: cachedStore,
			}
		default:
			return nil, fmt.Errorf("unsupported store type for commit: %T", commitStore)
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	finalizer := &multiTreeFinalizer{
		CommitMultiTree:    db,
		ctx:                ctx,
		cancel:             cancel,
		finalizers:         finalizers,
		workingCommitInfo:  commitInfo,
		done:               make(chan struct{}),
		hashReady:          make(chan struct{}),
		finalizeOrRollback: make(chan struct{}),
	}
	go func() {
		err := finalizer.commit(ctx)
		if err != nil {
			finalizer.err.Store(err)
		}
		close(finalizer.done)
	}()
	return finalizer, nil
}

type multiTreeFinalizer struct {
	*CommitMultiTree
	ctx                context.Context
	cancel             context.CancelFunc
	finalizers         []storetypes.CommitFinalizer
	workingCommitInfo  *storetypes.CommitInfo
	workingCommitId    storetypes.CommitID
	done               chan struct{}
	hashReady          chan struct{}
	finalizeOnce       sync.Once
	finalizeOrRollback chan struct{}
	err                atomic.Value
}

func (db *multiTreeFinalizer) commit(ctx context.Context) error {
	db.commitMutex.Lock()
	defer db.commitMutex.Unlock()

	stagedVersion := db.stagedVersion()
	ctx, span := tracer.Start(ctx, "CommitMultiTree.Commit",
		trace.WithAttributes(
			attribute.Int64("version", int64(stagedVersion)),
		),
	)
	defer span.End()

	if err := db.prepareCommit(ctx); err != nil {
		// rollback

		// do not use an errGroup here since context.Canceled is expected!
		var wg sync.WaitGroup
		var rbErr atomic.Value
		for _, finalizer := range db.finalizers {
			finalizer := finalizer
			wg.Add(1)
			go func() {
				defer wg.Done()
				// TODO check for errors in rollback that aren't context.Canceled
				err := finalizer.Rollback()
				if err != nil && !errors.Is(err, context.Canceled) {
					rbErr.Store(err)
				}
			}()
		}

		wg.Wait()
		if err := rbErr.Load(); err != nil {
			return fmt.Errorf("rollback failed: %w", err.(error))
		}

		return fmt.Errorf("successful rollback: %w; cause: %v", context.Canceled, err)
	}

	var errGroup errgroup.Group
	for _, finalizer := range db.finalizers {
		finalizer := finalizer
		errGroup.Go(func() error {
			_, err := finalizer.Finalize()
			return err
		})
	}
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("finalizing commit failed: %w", err)
	}

	err := saveCommitInfo(db.dir, db.stagedVersion(), db.workingCommitInfo)
	if err != nil {
		return fmt.Errorf("failed to save commit info for version %d: %v", db.stagedVersion(), err)
	}
	db.lastCommitId = db.workingCommitId
	db.lastCommitInfo = db.workingCommitInfo
	db.version++
	return nil
}

func (db *multiTreeFinalizer) prepareCommit(ctx context.Context) error {
	var hashErrGroup errgroup.Group
	for i, finalizer := range db.finalizers {
		finalizer := finalizer
		hashErrGroup.Go(func() error {
			hash, err := finalizer.WorkingHash()
			if err != nil {
				return err
			}
			db.workingCommitInfo.StoreInfos[i].CommitId = hash
			return nil
		})
	}
	if err := hashErrGroup.Wait(); err != nil {
		return err
	}

	db.workingCommitId = storetypes.CommitID{
		Version: int64(db.stagedVersion()),
		Hash:    db.workingCommitInfo.Hash(),
	}
	close(db.hashReady)

	<-db.finalizeOrRollback

	return ctx.Err()
}

func (db *multiTreeFinalizer) WorkingHash() (storetypes.CommitID, error) {
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

func (db *multiTreeFinalizer) SignalFinalize() error {
	db.finalizeOnce.Do(func() {
		close(db.finalizeOrRollback)
	})
	return nil
}

func (db *multiTreeFinalizer) Finalize() (storetypes.CommitID, error) {
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

func (db *multiTreeFinalizer) Rollback() error {
	db.cancel()
	close(db.finalizeOrRollback)
	<-db.done
	err := db.err.Load()
	if err != nil {
		// we expect an error if we rolled back successfully
		return err.(error)
	}
	return fmt.Errorf("expected error on rollback, got nil")
}

type fauxCommitStoreFinalizer struct {
	committer storetypes.Committer
	cacheWrap storetypes.CacheWrap
}

func (f *fauxCommitStoreFinalizer) WorkingHash() (storetypes.CommitID, error) {
	return storetypes.CommitID{}, nil
}

func (f *fauxCommitStoreFinalizer) SignalFinalize() error {
	return nil
}

func (f *fauxCommitStoreFinalizer) Finalize() (storetypes.CommitID, error) {
	if f.cacheWrap != nil {
		f.cacheWrap.Write()
	}
	// commit doesn't actually compute any hash here
	return f.committer.Commit(), nil
}

func (f *fauxCommitStoreFinalizer) Rollback() error {
	return nil
}

func (db *CommitMultiTree) GetCommitInfo(ver int64) (*storetypes.CommitInfo, error) {
	return loadCommitInfo(db.dir, uint64(ver))
}

func (db *CommitMultiTree) Commit() storetypes.CommitID {
	panic("cannot call Commit on uncached CommitMultiTree directly; use StartCommit")
}

func (db *CommitMultiTree) WorkingHash() []byte {
	panic("cannot call WorkingHash on uncached CommitMultiTree directly; use StartCommit")
}

func (db *CommitMultiTree) LastCommitID() storetypes.CommitID {
	return db.lastCommitId
}

const commitInfoSubPath = "commit_info"

// saveCommitInfo saves the CommitInfo for a given version to a <version>.ci file.
func saveCommitInfo(dir string, version uint64, commitInfo *storetypes.CommitInfo) error {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", version))
	bz, err := proto.Marshal(commitInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal commit info for version %d: %w", version, err)
	}

	file, err := os.OpenFile(commitInfoPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open commit info file for version %d: %w", version, err)
	}

	_, err = file.Write(bz)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to write commit info file for version %d: %w", version, err)
	}

	err = file.Sync()
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to sync commit info file for version %d: %w", version, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close commit info file for version %d: %w", version, err)
	}

	// TODO consider fsyncing the directory as well for extra durability guarantees

	return nil
}

// loadLatestCommitInfo loads the highest version number commit info file from the commit_info directory
// if any exist, returning the version and CommitInfo.
func loadLatestCommitInfo(dir string) (uint64, *storetypes.CommitInfo, error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err := os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	entries, err := os.ReadDir(commitInfoDir)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read commit info dir: %w", err)
	}

	// find the latest version by looking for the highest numbered file
	var latestVersion uint64
	for _, entry := range entries {
		var version uint64
		_, err := fmt.Sscanf(entry.Name(), "%d", &version)
		if err != nil {
			// skip non-numeric files
			continue
		}
		if version > latestVersion {
			latestVersion = version
		}
	}

	if latestVersion == 0 {
		// no versions found, no commit info to load
		return 0, nil, nil
	}

	commitInfo, err := loadCommitInfo(dir, latestVersion)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to load commit info for version %d: %w", latestVersion, err)
	}

	return latestVersion, commitInfo, nil
}

func loadCommitInfo(dir string, version uint64) (*storetypes.CommitInfo, error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err := os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", version))
	bz, err := os.ReadFile(commitInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit info file for version %d: %w", version, err)
	}

	commitInfo := &storetypes.CommitInfo{}
	err = proto.Unmarshal(bz, commitInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal commit info for version %d: %w", version, err)
	}

	return commitInfo, nil
}

func (db *CommitMultiTree) SetPruning(pruningtypes.PruningOptions) {
	logger.Warn("SetPruning is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) GetPruning() pruningtypes.PruningOptions {
	return pruningtypes.NewPruningOptions(pruningtypes.PruningDefault)
}

func (db *CommitMultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (db *CommitMultiTree) GetStore(storetypes.StoreKey) storetypes.Store {
	panic("cannot call GetStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) GetKVStore(storetypes.StoreKey) storetypes.KVStore {
	panic("cannot call GetKVStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

// GetObjKVStore returns a mounted ObjKVStore for a given StoreKey.
func (db *CommitMultiTree) GetObjKVStore(storetypes.StoreKey) storetypes.ObjKVStore {
	panic("cannot call GetObjKVStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) GetCommitStore(storetypes.StoreKey) storetypes.CommitStore {
	panic("cannot call GetCommitStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) GetCommitKVStore(storetypes.StoreKey) storetypes.CommitKVStore {
	panic("cannot call GetCommitKVStore on uncached CommitMultiTree directly; use CacheMultiStore first")
}

func (db *CommitMultiTree) SetTracer(io.Writer) storetypes.MultiStore {
	panic("SetTracer is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetTracingContext(storetypes.TraceContext) storetypes.MultiStore {
	panic("SetTracingContext is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) TracingEnabled() bool {
	return false
}

func (db *CommitMultiTree) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return fmt.Errorf("snapshotting has not been implemented yet")
}

func (db *CommitMultiTree) PruneSnapshotHeight(height int64) {
	logger.Warn("PruneSnapshotHeight is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetSnapshotInterval(snapshotInterval uint64) {
	logger.Warn("SetSnapshotInterval is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) Restore(height uint64, format uint32, protoReader protoio.Reader) (snapshottypes.SnapshotItem, error) {
	return snapshottypes.SnapshotItem{}, fmt.Errorf("restoring from snapshot has not been implemented yet")
}

func (db *CommitMultiTree) MountStoreWithDB(key storetypes.StoreKey, typ storetypes.StoreType, _ dbm.DB) {
	if _, exists := db.treesByKey[key]; exists {
		panic(fmt.Sprintf("store with key %s already mounted", key.Name()))
	}
	index := len(db.treeKeys)
	db.treesByKey[key] = index
	db.treeKeys = append(db.treeKeys, key)
	db.storeTypes = append(db.storeTypes, typ)
}

func (db *CommitMultiTree) LoadLatestVersion() error {
	for i, key := range db.treeKeys {
		storeType := db.storeTypes[i]
		tree, err := db.loadStore(key, storeType)
		if err != nil {
			return fmt.Errorf("failed to load store %s: %w", key.Name(), err)
		}
		db.trees = append(db.trees, tree)
	}

	version, ci, err := loadLatestCommitInfo(db.dir)
	if err != nil {
		return fmt.Errorf("failed to load latest commit info: %w", err)
	}

	db.version = version
	if ci != nil {
		// should be nil on initial creation
		db.lastCommitId = storetypes.CommitID{
			Version: int64(version),
			Hash:    ci.Hash(),
		}
		db.lastCommitInfo = ci
	}

	return nil
}

func (db *CommitMultiTree) loadStore(key storetypes.StoreKey, typ storetypes.StoreType) (storetypes.CacheWrapper, error) {
	switch typ {
	case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
		dir := filepath.Join(db.dir, "stores", fmt.Sprintf(key.Name(), ".iavl"))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create store dir %s: %w", dir, err)
			}
		}
		return NewCommitTree(dir, db.opts)
	case storetypes.StoreTypeTransient:
		_, ok := key.(*storetypes.TransientStoreKey)
		if !ok {
			return nil, fmt.Errorf("invalid StoreKey for StoreTypeTransient: %s", key.String())
		}

		return transient.NewStore(), nil
	case storetypes.StoreTypeMemory:
		if _, ok := key.(*storetypes.MemoryStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a MemoryStoreKey; got: %s", key.String())
		}

		return mem.NewStore(), nil
	case storetypes.StoreTypeObject:
		if _, ok := key.(*storetypes.ObjectStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a ObjectStoreKey: %s", key.String())
		}
		return transient.NewObjStore(), nil
	default:
		return nil, fmt.Errorf("unsupported store type: %s", typ.String())
	}
}

func (db *CommitMultiTree) LoadLatestVersionAndUpgrade(upgrades *storetypes.StoreUpgrades) error {
	return fmt.Errorf("LoadLatestVersionAndUpgrade has not been implemented yet")
}

func (db *CommitMultiTree) LoadVersionAndUpgrade(ver int64, upgrades *storetypes.StoreUpgrades) error {
	return fmt.Errorf("LoadVersionAndUpgrade has not been implemented yet")
}

func (db *CommitMultiTree) LoadVersion(ver int64) error {
	return fmt.Errorf("LoadVersion has not been implemented yet")
}

func (db *CommitMultiTree) SetInterBlockCache(cache storetypes.MultiStorePersistentCache) {
	logger.Warn("SetInterBlockCache is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetInitialVersion(version int64) error {
	return fmt.Errorf("SetInitialVersion has not been implemented yet")
}

func (db *CommitMultiTree) SetIAVLCacheSize(size int) {}

func (db *CommitMultiTree) SetIAVLDisableFastNode(disable bool) {}

func (db *CommitMultiTree) SetIAVLSyncPruning(sync bool) {}

func (db *CommitMultiTree) RollbackToVersion(version int64) error {
	return fmt.Errorf("RollbackToVersion has not been implemented yet")
}

func (db *CommitMultiTree) ListeningEnabled(key storetypes.StoreKey) bool {
	logger.Warn("ListeningEnabled is not implemented for CommitMultiTree")
	return false
}

func (db *CommitMultiTree) AddListeners(keys []storetypes.StoreKey) {
	logger.Warn("AddListeners is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) PopStateCache() []*storetypes.StoreKVPair {
	// TODO implement me
	panic("implement me")
}

func (db *CommitMultiTree) SetMetrics(metrics metrics.StoreMetrics) {
	logger.Warn("SetMetrics is not implemented for CommitMultiTree")
}

func LoadDB(path string, opts Options) (*CommitMultiTree, error) {
	db := &CommitMultiTree{
		dir:        path,
		opts:       opts,
		treesByKey: make(map[storetypes.StoreKey]int),
	}
	return db, nil
}

func (db *CommitMultiTree) stagedVersion() uint64 {
	return db.version + 1
}

func (db *CommitMultiTree) LatestVersion() int64 {
	return int64(db.version)
}

func (db *CommitMultiTree) Close() error {
	for _, tree := range db.trees {
		if closer, ok := tree.(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *CommitMultiTree) CacheWrap() storetypes.CacheWrap {
	return db.CacheMultiStore()
}

func (db *CommitMultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	logger.Warn("CacheWrapWithTrace is not implemented for CommitMultiTree; falling back to CacheWrap")
	return db.CacheWrap()
}

func (db *CommitMultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	return internal.NewMultiTree(int64(db.version), func(key storetypes.StoreKey) storetypes.CacheWrap {
		idx, ok := db.treesByKey[key]
		if !ok {
			panic(fmt.Sprintf("store with key %s not mounted", key.Name()))
		}
		tree := db.trees[idx]
		return tree.CacheWrap()
	})
}

func (db *CommitMultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	if version == 0 {
		// use latest version
		return db.CacheMultiStore(), nil
	}

	mt := internal.NewMultiTree(version, func(key storetypes.StoreKey) storetypes.CacheWrap {
		idx, ok := db.treesByKey[key]
		if !ok {
			panic(fmt.Sprintf("store with key %s not mounted", key.Name()))
		}
		tree := db.trees[idx]
		switch tree := tree.(type) {
		case *CommitTree:
			// it's not really ideal to panic here, but the MultiTree interface doesn't allow for error returns
			// alternatively we can check out all historical roots aggressively, but that may be unnecessary when
			// historical queries will often touch a single store
			// a better approach may be to keep some global tracking of historical roots in the CommitMultiTree itself
			// by pruning old commit_info files when pruning is implemented
			t, err := tree.GetVersion(version)
			if err != nil {
				panic(fmt.Sprintf("failed to get version %d for store %s: %v", version, key.Name(), err))
			}
			return t.CacheWrap()
		default:
			return tree.CacheWrap()
		}
	})

	return mt, nil
}

var _ storetypes.CommitMultiStore2 = &CommitMultiTree{}
