package iavlx

import (
	"bytes"
	"fmt"
	io "io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alitto/pond/v2"
	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"

	"cosmossdk.io/log"
	"cosmossdk.io/store/mem"
	"cosmossdk.io/store/metrics"
	pruningtypes "cosmossdk.io/store/pruning/types"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	"cosmossdk.io/store/transient"
	storetypes "cosmossdk.io/store/types"
)

type CommitMultiTree struct {
	dir        string
	opts       Options
	logger     log.Logger
	trees      []storetypes.CommitKVStore  // always ordered by tree name
	treeKeys   []storetypes.StoreKey       // always ordered by tree name
	storeTypes []storetypes.StoreType      // store types by tree index
	treesByKey map[storetypes.StoreKey]int // index of the trees by name

	version           uint64
	lastCommitId      storetypes.CommitID
	commitPool        pond.ResultPool[[]byte]
	workingCommitInfo *storetypes.CommitInfo
	workingHash       []byte
}

func (db *CommitMultiTree) LastCommitID() storetypes.CommitID {
	return db.lastCommitId
}

func (db *CommitMultiTree) WorkingHash() []byte {
	// NOTE: this may invoke some hash recomputation each time even if there is no change
	taskGroup := db.commitPool.NewGroup()
	stagedVersion := db.version + 1
	for _, tree := range db.trees {
		t := tree
		taskGroup.Submit(func() []byte {
			return t.WorkingHash()
		})
	}
	hashes, err := taskGroup.Wait()
	if err != nil {
		panic(fmt.Errorf("failed to commit trees: %w", err))
	}

	commitInfo := &storetypes.CommitInfo{}
	commitInfo.StoreInfos = make([]storetypes.StoreInfo, len(db.treeKeys))
	for i, treeKey := range db.treeKeys {
		commitInfo.StoreInfos[i] = storetypes.StoreInfo{
			Name: treeKey.Name(),
			CommitId: storetypes.CommitID{
				Version: int64(stagedVersion),
				Hash:    hashes[i],
			},
		}
	}
	db.workingCommitInfo = commitInfo
	hash := commitInfo.Hash()
	db.workingHash = hash
	return hash
}

func (db *CommitMultiTree) Commit() storetypes.CommitID {
	// NOTE: this function is maybe unnecessarily complex because the SDK has both WorkingHash and Commit methods
	// and we're trying to avoid recomputing the hash
	// so we check if we already have a hash that was computed in WorkingHash that hasn't changed to avoid recomputation
	// in the future we should evaluate if there is any need to retain both WorkingHash and Commit methods separately
	taskGroup := db.commitPool.NewGroup()
	for _, tree := range db.trees {
		t := tree
		taskGroup.Submit(func() []byte {
			commitId := t.Commit()
			return commitId.Hash
		})
	}

	hashes, err := taskGroup.Wait()
	if err != nil {
		panic(fmt.Errorf("failed to commit trees: %w", err))
	}

	stagedVersion := db.version + 1
	commitInfo := db.workingCommitInfo
	var hash []byte
	if commitInfo == nil {
		commitInfo = &storetypes.CommitInfo{}
		commitInfo.StoreInfos = make([]storetypes.StoreInfo, len(db.treeKeys))
		for i, treeKey := range db.treeKeys {
			commitInfo.StoreInfos[i] = storetypes.StoreInfo{
				Name: treeKey.Name(),
				CommitId: storetypes.CommitID{
					Version: int64(stagedVersion),
					Hash:    hashes[i],
				},
			}
		}
		hash = commitInfo.Hash()
	} else {
		hashChanged := false
		for i, storeInfo := range commitInfo.StoreInfos {
			if !bytes.Equal(storeInfo.CommitId.Hash, hashes[i]) {
				hashChanged = true
				commitInfo.StoreInfos[i].CommitId.Hash = hashes[i]
			}
		}
		if !hashChanged {
			hash = db.workingHash
		} else {
			hash = commitInfo.Hash()
		}
		db.workingCommitInfo = nil
		db.workingHash = nil
	}

	db.version++
	commitId := storetypes.CommitID{
		Version: int64(db.version),
		Hash:    hash,
	}
	db.lastCommitId = commitId
	return commitId
}

func (db *CommitMultiTree) SetPruning(options pruningtypes.PruningOptions) {
	db.logger.Warn("SetPruning is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) GetPruning() pruningtypes.PruningOptions {
	return pruningtypes.NewPruningOptions(pruningtypes.PruningDefault)
}

func (db *CommitMultiTree) GetStoreType() storetypes.StoreType {
	return storetypes.StoreTypeMulti
}

func (db *CommitMultiTree) CacheWrap() storetypes.CacheWrap {
	return db.CacheMultiStore()
}

func (db *CommitMultiTree) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	// TODO implement tracking
	return db.CacheMultiStore()
}

func (db *CommitMultiTree) CacheMultiStore() storetypes.CacheMultiStore {
	mt := &MultiTree{
		trees:      make([]storetypes.CacheKVStore, len(db.trees)),
		treesByKey: db.treesByKey, // share the map
	}
	for i, tree := range db.trees {
		mt.trees[i] = tree.CacheWrap().(storetypes.CacheKVStore)
	}
	return mt
}

func (db *CommitMultiTree) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	if version == 0 {
		version = int64(db.version)
	}

	mt := &MultiTree{
		latestVersion: version,
		treesByKey:    db.treesByKey, // share the map
		trees:         make([]storetypes.CacheKVStore, len(db.trees)),
	}

	for i, tree := range db.trees {
		typ := db.storeTypes[i]
		switch typ {
		case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
			var err error
			mt.trees[i], err = tree.(*CommitTree).GetImmutable(version)
			if err != nil {
				return nil, fmt.Errorf("failed to create cache multi store for tree %s at version %d: %w", db.treeKeys[i].Name(), version, err)
			}
		case storetypes.StoreTypeTransient, storetypes.StoreTypeMemory:
			mt.trees[i] = tree.CacheWrap().(storetypes.CacheKVStore)
		default:
			return nil, fmt.Errorf("unsupported store type: %s", typ.String())
		}
	}

	return mt, nil
}

func (db *CommitMultiTree) GetStore(key storetypes.StoreKey) storetypes.Store {
	return db.trees[db.treesByKey[key]]
}

func (db *CommitMultiTree) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	index, ok := db.treesByKey[key]
	if !ok {
		panic(fmt.Sprintf("store not found for key: %s (key type: %T)", key.Name(), key))
	}
	if index >= len(db.trees) {
		panic(fmt.Sprintf("store index %d out of bounds for key %s (trees length: %d)", index, key.Name(), len(db.trees)))
	}
	return db.trees[index]
}

func (db *CommitMultiTree) TracingEnabled() bool {
	return false
}

func (db *CommitMultiTree) SetTracer(w io.Writer) storetypes.MultiStore {
	db.logger.Warn("SetTracer is not implemented for CommitMultiTree")
	return db
}

func (db *CommitMultiTree) SetTracingContext(context storetypes.TraceContext) storetypes.MultiStore {
	db.logger.Warn("SetTracingContext is not implemented for CommitMultiTree")
	return db
}

func (db *CommitMultiTree) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return fmt.Errorf("snapshotting has not been implemented yet")
}

func (db *CommitMultiTree) PruneSnapshotHeight(height int64) {
	db.logger.Warn("PruneSnapshotHeight is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetSnapshotInterval(snapshotInterval uint64) {
	db.logger.Warn("SetSnapshotInterval is not implemented for CommitMultiTree")
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

func (db *CommitMultiTree) GetCommitStore(key storetypes.StoreKey) storetypes.CommitStore {
	return db.trees[db.treesByKey[key]]
}

func (db *CommitMultiTree) GetCommitKVStore(key storetypes.StoreKey) storetypes.CommitKVStore {
	return db.trees[db.treesByKey[key]]
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
	return nil
}

func (db *CommitMultiTree) loadStore(key storetypes.StoreKey, typ storetypes.StoreType) (storetypes.CommitKVStore, error) {
	switch typ {
	case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
		dir := filepath.Join(db.dir, key.Name())
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			return nil, fmt.Errorf("store directory %s already exists, reloading isn't supported yet", dir)
		}
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return nil, fmt.Errorf("failed to create store dir %s: %w", dir, err)
		}
		return NewCommitTree(dir, db.opts, db.logger.With("store", key.Name()))
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
	db.logger.Warn("SetInterBlockCache is not implemented for CommitMultiTree")
}

func (db *CommitMultiTree) SetInitialVersion(version int64) error {
	return fmt.Errorf("SetInitialVersion has not been implemented yet")
}

func (db *CommitMultiTree) SetIAVLCacheSize(size int) {
}

func (db *CommitMultiTree) SetIAVLDisableFastNode(disable bool) {
}

func (db *CommitMultiTree) SetIAVLSyncPruning(sync bool) {
}

func (db *CommitMultiTree) RollbackToVersion(version int64) error {
	return fmt.Errorf("RollbackToVersion has not been implemented yet")
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

func (db *CommitMultiTree) SetMetrics(metrics metrics.StoreMetrics) {
	db.logger.Warn("SetMetrics is not implemented for CommitMultiTree")
}

func LoadDB(path string, opts *Options, logger log.Logger) (*CommitMultiTree, error) {
	// n := len(treeNames)
	//trees := make([]*CommitTree, n)
	//treesByName := make(map[string]int, n)
	//for i, name := range treeNames {
	//	if _, exists := treesByName[name]; exists {
	//		return nil, fmt.Errorf("duplicate tree name: %s", name)
	//	}
	//	treesByName[name] = i
	//	dir := filepath.Join(path, name)
	//	err := os.MkdirAll(dir, 0o755)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to create tree dir %s: %w", dir, err)
	//	}
	//	// Create a logger with tree name context
	//	treeLogger := logger.With("tree", name)
	//	trees[i], err = NewCommitTree(dir, *opts, treeLogger)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to load tree %s: %w", name, err)
	//	}
	//}
	//
	db := &CommitMultiTree{
		dir:        path,
		opts:       *opts,
		commitPool: pond.NewResultPool[[]byte](runtime.NumCPU()),
		logger:     logger,
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
		return nil
	}
	return nil
}

var _ storetypes.CommitMultiStore = &CommitMultiTree{}
