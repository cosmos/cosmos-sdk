package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	errorsmod "cosmossdk.io/errors"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/store/v2/mem"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
	snapshottypes "github.com/cosmos/cosmos-sdk/store/v2/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store/v2/transient"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

// CommitMultiTree implements storetypes.CommitMultiStore for iavlx.
//
// Implementation status of the CommitMultiStore interface:
//
// Fully implemented:
//   - Commit path: CacheMultiStore, CacheMultiStoreWithVersion, CommitBranch, WorkingHash, Commit
//   - Read path: GetCommitInfo, LastCommitID, LatestVersion, EarliestVersion, Query
//   - Lifecycle: MountStoreWithDB, LoadLatestVersion, Close
//   - Pruning: SetPruning, GetPruning (with background compaction)
//   - Rollback: RollbackToVersion (performs filesystem rollback, then poisons in-memory state;
//     requires restart — see also the offline `iavlx rollback` CLI tool)
//
// Not yet implemented:
//   - Snapshots / state sync: Snapshot, Restore, PruneSnapshotHeight, SetSnapshotInterval
//     (export.go and importer.go have the per-tree primitives, but they are not wired into
//     the CommitMultiStore snapshot interface yet)
//   - Store upgrades: LoadLatestVersionAndUpgrade, LoadVersionAndUpgrade, LoadVersion
//   - SetInitialVersion
//   - Listeners: ListeningEnabled, AddListeners, PopStateCache
//
// Import from iavl/v1:
//   The `iavlx import` CLI tool can import an existing iavl/v1 LevelDB-backed multi-store
//   into iavlx format. See ImportIAVLV1MultiStore in import_v1.go and the Importer type
//   in importer.go. This is a one-time offline migration, not part of the CommitMultiStore interface.
var (
	_ storetypes.CommitMultiStore = &CommitMultiTree{}
	_ storetypes.Queryable        = &CommitMultiTree{}
)

type CommitMultiTree struct {
	dir         string
	opts        Options
	stores      []*storeData                // always ordered by name
	iavlStores  []*storeData                // subset of stores that are IAVL, ordered by name
	otherStores []*storeData                // subset of stores that are not IAVL
	storesByKey map[storetypes.StoreKey]int // index of the stores by name

	commitMutex     sync.Mutex
	commitData      atomic.Pointer[commitData]
	earliestVersion atomic.Int64

	compactionMu     sync.Mutex
	cancelCompaction context.CancelFunc
	compactionActive atomic.Bool
	compactionDone   chan struct{}
	pruningOptions   pruningtypes.PruningOptions
	logger           log.Logger

	// compatFinalizer is a HACK to make iavlx work with the existing store/v1 and store/v2 CommitMultiStore interfaces.
	// Essentially we have a way of tracking that someone called Write on the root CacheMultiStore and started a
	// background commit. With this compatFinalizer type we then can allow the caller to call WorkingHash() and
	// Commit() as if this were a v1 or v2 store.
	compatFinalizer   *CommitFinalizer
	compatFinalizerMu sync.RWMutex
}

type commitData struct {
	commitInfo *storetypes.CommitInfo
	commitId   storetypes.CommitID
}

type storeData struct {
	key   storetypes.StoreKey
	typ   storetypes.StoreType
	store any
}

func (db *CommitMultiTree) GetCommitInfo(ver int64) (*storetypes.CommitInfo, error) {
	return loadCommitInfo(db.dir, ver)
}

func (db *CommitMultiTree) CommitBranch() *CommitBranch {
	return &CommitBranch{
		MultiTree: db.rootCacheMultiStore(),
		db:        db,
	}
}

func (db *CommitMultiTree) EarliestVersion() int64 {
	return db.earliestVersion.Load()
}

func (db *CommitMultiTree) LastCommitID() storetypes.CommitID {
	cd := db.commitData.Load()
	if cd == nil {
		return storetypes.CommitID{}
	}
	return cd.commitId
}

func (db *CommitMultiTree) SetPruning(opts pruningtypes.PruningOptions) {
	db.pruningOptions = opts
}

func (db *CommitMultiTree) GetPruning() pruningtypes.PruningOptions {
	return db.pruningOptions
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

// Snapshot is not yet wired up to the CometBFT state sync protocol.
// The per-tree primitives exist: TreeReader.Export (export.go) for serialization, Importer
// (importer.go) for deserialization. What remains is the multi-store orchestration and the
// snapshot framing format.
func (db *CommitMultiTree) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return fmt.Errorf("snapshotting has not been implemented yet")
}

func (db *CommitMultiTree) PruneSnapshotHeight(height int64) {
}

func (db *CommitMultiTree) SetSnapshotInterval(snapshotInterval uint64) {
}

// Restore is not yet wired up. See Snapshot.
func (db *CommitMultiTree) Restore(height uint64, format uint32, protoReader protoio.Reader) (snapshottypes.SnapshotItem, error) {
	return snapshottypes.SnapshotItem{}, fmt.Errorf("restoring from snapshot has not been implemented yet")
}

func (db *CommitMultiTree) MountStoreWithDB(key storetypes.StoreKey, typ storetypes.StoreType, _ dbm.DB) {
	if _, exists := db.storesByKey[key]; exists {
		panic(fmt.Sprintf("store with key %s already mounted", key.Name()))
	}
	db.storesByKey[key] = -1 // we assign actual index when loading
	db.stores = append(db.stores, &storeData{
		key: key,
		typ: typ,
	})
}

// LoadLatestVersion initializes all mounted stores by loading their on-disk data.
// It reads the latest commit info to determine the expected version, then loads each
// store (creating directories as needed for new stores). Must be called after all
// MountStoreWithDB calls and before any reads or commits.
func (db *CommitMultiTree) LoadLatestVersion() error {
	// sort treeKeys to ensure deterministic order
	// we assume that MountStoreWithDB has been called for all stores before this
	// so treeKeys and storeTypes are aligned by index
	slices.SortFunc(db.stores, func(a, b *storeData) int {
		return bytes.Compare([]byte(a.key.Name()), []byte(b.key.Name()))
	})

	ci, earliestVersion, err := loadLatestCommitInfo(db.dir)
	if err != nil {
		return fmt.Errorf("failed to load latest commit info: %w", err)
	}

	var version int64
	if ci != nil {
		// should be nil on initial creation
		version = ci.Version
		db.earliestVersion.Store(earliestVersion)
	}

	for i, si := range db.stores {
		key := si.key
		storeType := si.typ
		store, err := db.loadStore(si.key, storeType, uint64(version))
		if err != nil {
			return fmt.Errorf("failed to load store %s: %w", key.Name(), err)
		}
		si.store = store
		db.storesByKey[key] = i
		if _, ok := store.(*CommitTree); ok {
			db.iavlStores = append(db.iavlStores, si)
		} else {
			db.otherStores = append(db.otherStores, si)
		}
	}

	if ci != nil {
		// The commit info hash footer is not fsynced, so after a crash per-store
		// hashes may be missing. Recompute any empty ones from the loaded trees
		// so the app hash matches what CometBFT expects.
		for i, si := range ci.StoreInfos {
			// If we have a valid hash, no need to recompute.
			if len(si.CommitId.Hash) != 0 {
				continue
			}
			// We don't have the store key in StoreInfos, so we need to search all iavl stores.
			for _, sd := range db.iavlStores {
				if sd.key.Name() == si.Name {
					_, rootPtr := sd.store.(*CommitTree).treeStore.Latest()
					hash, err := rootHash(context.Background(), rootPtr)
					if err != nil {
						return fmt.Errorf("recomputing root hash for store %s: %w", si.Name, err)
					}
					ci.StoreInfos[i].CommitId = storetypes.CommitID{Version: ci.Version, Hash: hash}
					break
				}
			}
		}
		db.commitData.Store(&commitData{
			commitId:   storetypes.CommitID{Version: version, Hash: ci.Hash()},
			commitInfo: ci,
		})
	}

	return nil
}

func (db *CommitMultiTree) loadStore(key storetypes.StoreKey, typ storetypes.StoreType, expectedVersion uint64) (any, error) {
	switch typ {
	case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
		dir := filepath.Join(db.dir, "stores", fmt.Sprintf("%s.iavl", key.Name()))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create store dir %s: %w", dir, err)
			}
		}
		ct, err := NewCommitTree(dir, TreeOptions{
			Options:         db.opts,
			TreeName:        key.Name(),
			ExpectedVersion: uint32(expectedVersion),
		}, db.logger.With("store", key.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to load CommitTree for store %s: %w", key.Name(), err)
		}
		if uint64(ct.LatestVersion()) != expectedVersion {
			return nil, fmt.Errorf("store %s version mismatch: expected %d, got %d", key.Name(), expectedVersion, ct.LatestVersion())
		}
		return ct, nil
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

func (db *CommitMultiTree) SetInitialVersion(version int64) error {
	return fmt.Errorf("SetInitialVersion has not been implemented yet")
}

// RollbackToVersion performs a filesystem-level rollback and then POISONS this CommitMultiTree.
//
// After this method returns, the in-memory state is intentionally destroyed (stores set to nil).
// Any subsequent use of this CommitMultiTree will panic. The caller MUST restart the process
// after rollback — there is no way to continue using a rolled-back store in the same process.
//
// This is by design: live rollback is an extremely rare operator-initiated action (not something
// that happens during normal consensus), and the complexity of safely reinitializing all in-memory
// state is not worth it. The node should be restarted after rollback, at which point the normal
// load() path will reconstruct the trees at the target version from the rolled-back files on disk.
//
// For a fully offline rollback (without loading the store first), use the `iavlx rollback` CLI tool.
func (db *CommitMultiTree) RollbackToVersion(version int64) error {
	db.commitMutex.Lock()
	defer db.commitMutex.Unlock()

	// Close all stores first so we release file handles, mmaps, etc.
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close stores before rollback: %w", err)
	}

	// Perform the filesystem-level rollback (WAL truncation, checkpoint rollback, commit info cleanup).
	if err := RollbackMultiTree(db.dir, uint64(version), db.logger, ""); err != nil {
		return fmt.Errorf("failed to rollback to version %d: %w", version, err)
	}

	// Poison the in-memory state. Setting stores to nil ensures any subsequent use
	// of this CommitMultiTree will panic with a nil pointer dereference, making it
	// obvious that the process needs to be restarted.
	db.stores = nil
	db.iavlStores = nil
	db.otherStores = nil
	db.storesByKey = nil

	db.logger.Info("rollback complete — the node MUST be restarted", "targetVersion", version)
	return nil
}

func (db *CommitMultiTree) ListeningEnabled(key storetypes.StoreKey) bool {
	db.logger.Warn("ListeningEnabled is not implemented for CommitMultiTree")
	return false
}

func (db *CommitMultiTree) AddListeners(keys []storetypes.StoreKey) {
	db.logger.Warn("AddListeners is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) PopStateCache() []*storetypes.StoreKVPair {
	// TODO implement me
	panic("implement me")
}

// LoadCommitMultiTree creates a new CommitMultiTree rooted at path. The caller must then
// mount stores via MountStoreWithDB and call LoadLatestVersion to initialize them.
func LoadCommitMultiTree(path string, opts Options, logger log.Logger) (*CommitMultiTree, error) {
	db := &CommitMultiTree{
		dir:            path,
		opts:           opts,
		storesByKey:    make(map[storetypes.StoreKey]int),
		pruningOptions: pruningtypes.NewPruningOptions(pruningtypes.PruningNothing),
		logger:         logger,
	}
	return db, nil
}

func (db *CommitMultiTree) stagedVersion() int64 {
	return db.LatestVersion() + 1
}

func (db *CommitMultiTree) LatestVersion() int64 {
	cd := db.commitData.Load()
	if cd == nil {
		return 0
	}
	return cd.commitId.Version
}

// Close cancels any in-progress compaction, waits for it to finish, and closes all stores.
func (db *CommitMultiTree) Close() error {
	db.compactionMu.Lock()
	cancel := db.cancelCompaction
	done := db.compactionDone
	db.compactionMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		// wait for any ongoing compaction to finish before closing stores
		<-done
	}
	var errGroup errgroup.Group
	for _, si := range db.stores {
		if closer, ok := si.store.(io.Closer); ok {
			errGroup.Go(closer.Close)
		}
	}
	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf("failed to close stores: %w", err)
	}
	return nil
}

func (db *CommitMultiTree) CacheWrap() storetypes.CacheWrap {
	panic("TODO")
}

func (db *CommitMultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	cb := db.CommitBranch()
	// Here we wrap our CommitBranch in a wrapper which allows the CacheMultiStore.Write()
	// method be used as a workaround for store/v1 and store/v2 compatibility.
	return &commitBranchCacheCompatWrapper{CommitBranch: cb}
}

func (db *CommitMultiTree) rootCacheMultiStore() *MultiTree {
	cd := db.commitData.Load()
	if cd == nil {
		return db.cacheMultiStore(0)
	}
	return db.cacheMultiStore(cd.commitId.Version)
}

func (db *CommitMultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	if version == 0 {
		// use latest version
		return db.CacheMultiStore(), nil
	}
	// check if we actually have CommitInfo for this version - basically fail fast when we don't
	_, err := loadCommitInfo(db.dir, version)
	if err != nil {
		return nil, fmt.Errorf("version %d is not available: %w", version, err)
	}

	return db.cacheMultiStore(version), nil
}

func (db *CommitMultiTree) cacheMultiStore(version int64) *MultiTree {
	return NewMultiTree(version, func(key storetypes.StoreKey) storetypes.CacheWrap {
		idx, ok := db.storesByKey[key]
		if !ok {
			panic(fmt.Sprintf("store with key %s not mounted", key.Name()))
		}
		tree := db.stores[idx].store
		switch tree := tree.(type) {
		case *CommitTree:
			// it's not really ideal to panic here, but the MultiTree interface doesn't allow for error returns
			// alternatively we can check out all historical roots aggressively, but that may be unnecessary when
			// historical queries will often touch a single store
			// a better approach may be to keep some global tracking of historical roots in the CommitMultiTree itself
			// by pruning old commit_info files when pruning is implemented
			t, err := tree.GetVersion(uint32(version))
			if err != nil {
				panic(fmt.Sprintf("failed to get version %d for store %s: %v", version, key.Name(), err))
			}
			return t.CacheWrap()
		case storetypes.CacheWrapper:
			return tree.CacheWrap()
		default:
			panic(fmt.Sprintf("store %s of type %T does not support caching", key.Name(), tree))
		}
	})
}

func (db *CommitMultiTree) WorkingHash() []byte {
	db.compatFinalizerMu.RLock()
	defer db.compatFinalizerMu.RUnlock()

	if db.compatFinalizer == nil {
		panic("no commit in progress")
	}
	cid, err := db.compatFinalizer.StartFinalize()
	if err != nil {
		panic(err)
	}
	return cid.Hash
}

func (db *CommitMultiTree) Commit() storetypes.CommitID {
	db.compatFinalizerMu.Lock()
	defer db.compatFinalizerMu.Unlock()

	if db.compatFinalizer == nil {
		panic("no commit in progress")
	}
	cid, err := db.compatFinalizer.Finalize()
	if err != nil {
		panic(err)
	}

	// Clear the finalizer to allow another commit to happen in the future.
	db.compatFinalizer = nil

	return cid
}

func (db *CommitMultiTree) SetInterBlockCache(cache storetypes.MultiStorePersistentCache) {
	db.logger.Warn("SetInterBlockCache is a no-op for iavlx")
}

func (db *CommitMultiTree) SetIAVLCacheSize(int) {
	db.logger.Warn("SetIAVLCacheSize is a no-op for iavlx")
}

func (db *CommitMultiTree) SetIAVLDisableFastNode(bool) {
	db.logger.Warn("SetIAVLDisableFastNode is a no-op for iavlx")
}

func (db *CommitMultiTree) SetIAVLSyncPruning(bool) {
	db.logger.Warn("SetIAVLSyncPruning is a no-op for iavlx")
}

func (db *CommitMultiTree) compactIfNeeded() {
	if db.pruningOptions.Strategy == pruningtypes.PruningNothing || db.pruningOptions.Interval == 0 {
		return
	}

	version := uint64(db.LatestVersion())
	if version%db.pruningOptions.Interval != 0 {
		return
	}

	// TODO: add snapshot awareness — when state sync snapshots are enabled, retainVersion
	// must not go below the oldest in-flight or most recent completed snapshot height.
	// See store/pruningmanager.go GetPruningHeight for reference.

	// Keep current version + KeepRecent previous versions
	if version <= db.pruningOptions.KeepRecent+1 {
		return
	}
	retainVersion := version - db.pruningOptions.KeepRecent

	if !db.compactionActive.CompareAndSwap(false, true) {
		// another compaction started since we checked, skip
		db.logger.Warn("skipping compaction since another compaction is already in progress",
			"version", version,
			"retain_version", retainVersion,
			"pruning_interval", db.pruningOptions.Interval,
			"keep_recent", db.pruningOptions.KeepRecent,
		)
		return
	}
	db.logger.Info("starting compaction of old versions of IAVL trees",
		"version", version,
		"retain_version", retainVersion,
		"pruning_interval", db.pruningOptions.Interval,
		"keep_recent", db.pruningOptions.KeepRecent,
	)

	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	db.compactionMu.Lock()
	db.compactionDone = done
	db.cancelCompaction = cancel
	db.compactionMu.Unlock()

	go func() {
		defer db.compactionActive.Store(false)
		defer close(done)
		db.compactNow(ctx, retainVersion)
	}()
}

// compactNow compacts old versions of the IAVL trees according to the pruning options, keeping the most recent `keepRecent` versions and compacting the rest.
// This function is only intended to be called from compactIfNeeded or by tests.
func (db *CommitMultiTree) compactNow(ctx context.Context, retainVersion uint64) {
	ctx, span := tracer.Start(ctx, "CommitMultiTree.Compact",
		trace.WithAttributes(
			attribute.Int64("retain_version", int64(retainVersion)),
			attribute.Int64("current_version", db.LatestVersion()),
		),
	)
	defer span.End()

	db.earliestVersion.Store(int64(retainVersion))

	err := deleteOldCommitInfos(db.dir, retainVersion)
	if err != nil {
		db.logger.Error(fmt.Sprintf("failed to delete old commit info files: %v", err))
	}

	for _, si := range db.iavlStores {
		ctx, span := tracer.Start(ctx, "CompactStore",
			trace.WithAttributes(
				attribute.String("store", si.key.Name()),
			),
		)
		ct, ok := si.store.(*CommitTree)
		if !ok {
			db.logger.Error(fmt.Sprintf("store %s is not a CommitTree, cannot compact", si.key.Name()))
			continue
		}
		err := ct.compact(ctx, uint32(retainVersion))
		if err != nil {
			db.logger.Error(fmt.Sprintf("failed to compact store %s: %v", si.key.Name(), err))
			continue
		}
		span.End()
	}
}

func deleteOldCommitInfos(dir string, retainVersion uint64) error {
	return deleteCommitInfos(dir, func(version uint64) bool {
		return version >= retainVersion
	})
}

func deleteCommitInfos(multiTreeDir string, retain func(uint64) bool) error {
	commitInfoDir := filepath.Join(multiTreeDir, commitInfoSubPath)
	entries, err := os.ReadDir(commitInfoDir)
	if err != nil {
		return fmt.Errorf("failed to read commit info dir: %w", err)
	}

	for _, entry := range entries {
		var version uint64
		_, err := fmt.Sscanf(entry.Name(), "%d", &version)
		if err != nil {
			continue
		}
		if !retain(version) {
			err := os.Remove(filepath.Join(commitInfoDir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to delete old commit info file %s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

// Query routes a query request to a sub-store by name and appends the multi-store proof when requested.
func (db *CommitMultiTree) Query(req *storetypes.RequestQuery) (*storetypes.ResponseQuery, error) {
	storeName, subpath, err := parseQueryPath(req.Path)
	if err != nil {
		return &storetypes.ResponseQuery{}, err
	}

	store := db.getStoreByName(storeName)
	if store == nil {
		return &storetypes.ResponseQuery{}, errorsmod.Wrapf(storetypes.ErrUnknownRequest, "no such store: %s", storeName)
	}

	queryable, ok := store.(storetypes.Queryable)
	if !ok {
		return &storetypes.ResponseQuery{}, errorsmod.Wrapf(storetypes.ErrUnknownRequest, "store %s (type %T) doesn't support queries", storeName, store)
	}

	subReq := *req
	subReq.Path = subpath
	res, err := queryable.Query(&subReq)
	if err != nil {
		return res, err
	}

	if !req.Prove || !queryRequiresProof(subpath) {
		return res, nil
	}

	if res.ProofOps == nil || len(res.ProofOps.Ops) == 0 {
		return &storetypes.ResponseQuery{}, errorsmod.Wrap(storetypes.ErrInvalidRequest, "proof is unexpectedly empty; ensure height has not been compacted away")
	}

	commitInfo, err := db.commitInfoForProof(res.Height)
	if err != nil {
		return &storetypes.ResponseQuery{}, err
	}
	if err := validateCommitInfoHash(commitInfo, storeName); err != nil {
		return &storetypes.ResponseQuery{}, err
	}

	res.ProofOps.Ops = append(res.ProofOps.Ops, commitInfo.ProofOp(storeName))

	return res, nil
}

func queryRequiresProof(subpath string) bool {
	return subpath == "/key"
}

func parseQueryPath(path string) (storeName, subpath string, err error) {
	if !strings.HasPrefix(path, "/") {
		return "", "", errorsmod.Wrapf(storetypes.ErrUnknownRequest, "invalid path: %s", path)
	}

	storeName, subpath, found := strings.Cut(path[1:], "/")
	if !found {
		return storeName, "", nil
	}

	return storeName, "/" + subpath, nil
}

func (db *CommitMultiTree) getStoreByName(name string) any {
	for _, si := range db.stores {
		if si.key.Name() == name {
			return si.store
		}
	}
	return nil
}

func (db *CommitMultiTree) commitInfoForProof(height int64) (*storetypes.CommitInfo, error) {
	lastCommitInfo := db.commitData.Load()
	latest := lastCommitInfo.commitId.Version
	if height == latest && lastCommitInfo != nil && lastCommitInfo.commitInfo.Version == height {
		return lastCommitInfo.commitInfo, nil
	}

	return db.GetCommitInfo(height)
}

func validateCommitInfoHash(commitInfo *storetypes.CommitInfo, storeName string) error {
	for _, storeInfo := range commitInfo.StoreInfos {
		if storeInfo.Name != storeName {
			continue
		}

		// TODO: we can actually recover here, but we'd need to recompute the hashes in the tree.
		if len(storeInfo.CommitId.Hash) == 0 {
			return errorsmod.Wrapf(storetypes.ErrInvalidRequest, "proof store hash is missing for store %s at height %d", storeName, commitInfo.Version)
		}

		return nil
	}

	return errorsmod.Wrapf(storetypes.ErrInvalidRequest, "proof store %s is missing from commit info at height %d", storeName, commitInfo.Version)
}
