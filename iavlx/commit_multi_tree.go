package iavlx

import (
	"bytes"
	"fmt"
	io "io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	dbm "github.com/cosmos/cosmos-db"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"

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
	logger     *slog.Logger
	trees      []storetypes.CommitStore    // always ordered by tree name
	treeKeys   []storetypes.StoreKey       // always ordered by tree name
	storeTypes []storetypes.StoreType      // store types by tree index
	treesByKey map[storetypes.StoreKey]int // index of the trees by name

	version           uint64
	lastCommitId      storetypes.CommitID
	workingCommitInfo *storetypes.CommitInfo
	workingHash       []byte
}

// GetObjKVStore returns a mounted ObjKVStore for a given StoreKey.
func (db *CommitMultiTree) GetObjKVStore(key storetypes.StoreKey) storetypes.ObjKVStore {
	treeIdx, ok := db.treesByKey[key]
	if !ok {
		panic(fmt.Sprintf("tree key not found in treesByKey: %v", key))
	}
	s := db.trees[treeIdx]
	if s == nil {
		panic(fmt.Sprintf("store does not exist for key: %s", key.Name()))
	}
	store, ok := s.(storetypes.ObjKVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not ObjKVStore", key))
	}

	return store
}

func (db *CommitMultiTree) LastCommitID() storetypes.CommitID {
	return db.lastCommitId
}

func (db *CommitMultiTree) WorkingHash() []byte {
	// NOTE: this may invoke some hash recomputation each time even if there is no change
	stagedVersion := db.version + 1
	hashes := make([][]byte, len(db.trees))
	var wg sync.WaitGroup
	for i, tree := range db.trees {
		wg.Add(1)
		go func(i int, t storetypes.CommitStore) {
			defer wg.Done()
			hashes[i] = t.WorkingHash()
		}(i, tree)
	}
	wg.Wait()

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
	hashes := make([][]byte, len(db.trees))
	var wg sync.WaitGroup
	for i, tree := range db.trees {
		wg.Add(1)
		go func(i int, t storetypes.CommitStore) {
			defer wg.Done()
			hashes[i] = t.Commit().Hash
		}(i, tree)
	}
	wg.Wait()

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

	err := saveCommitInfo(db.dir, stagedVersion, commitInfo)
	if err != nil {
		panic(fmt.Sprintf("failed to save commit info for version %d: %v", stagedVersion, err))
	}

	db.version++
	commitId := storetypes.CommitID{
		Version: int64(db.version),
		Hash:    hash,
	}
	db.lastCommitId = commitId
	return commitId
}

const commitInfoSubPath = "commit_info"
const latestFilename = "latest"

// saveCommitInfo saves the CommitInfo for a given version to a <version>.ci file,
// and updates the latest.version file to point to the latest version.
func saveCommitInfo(dir string, version uint64, commitInfo *storetypes.CommitInfo) error {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", version))
	bz, err := proto.Marshal(commitInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal commit info for version %d: %w", version, err)
	}
	err = os.WriteFile(commitInfoPath, bz, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write commit info file for version %d: %w", version, err)
	}

	latestVersionPath := filepath.Join(commitInfoDir, latestFilename)
	err = os.WriteFile(latestVersionPath, []byte(fmt.Sprintf("%d", version)), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write latest version file: %w", err)
	}

	return nil
}

// loadLatestCommitInfo loads the latest.version file to determine the latest version,
// and then loads the <version>.ci file to get the CommitInfo for that version.
func loadLatestCommitInfo(dir string) (uint64, *storetypes.CommitInfo, error) {
	commitInfoDir := filepath.Join(dir, commitInfoSubPath)
	err := os.MkdirAll(commitInfoDir, 0o700)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create commit info dir: %w", err)
	}

	latestVersionPath := filepath.Join(commitInfoDir, latestFilename)
	if _, err := os.Stat(latestVersionPath); os.IsNotExist(err) {
		return 0, nil, nil
	}

	bz, err := os.ReadFile(latestVersionPath)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read latest version file: %w", err)
	}
	var version uint64
	_, err = fmt.Sscanf(string(bz), "%d", &version)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse latest version: %w", err)
	}
	commitInfoPath := filepath.Join(commitInfoDir, fmt.Sprintf("%d", version))
	bz, err = os.ReadFile(commitInfoPath)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to read commit info file for version %d: %w", version, err)
	}

	commitInfo := &storetypes.CommitInfo{}
	err = proto.Unmarshal(bz, commitInfo)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to unmarshal commit info for version %d: %w", version, err)
	}

	return version, commitInfo, nil
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
		trees:      make([]storetypes.CacheWrap, len(db.trees)),
		treesByKey: db.treesByKey, // share the map
	}
	for i, tree := range db.trees {
		mt.trees[i] = tree.CacheWrap()
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
		trees:         make([]storetypes.CacheWrap, len(db.trees)),
	}

	// TODO: we should actually use the CommitInfo for this version to load the correct set of trees (because there may have been store additions/removals)
	for i, tree := range db.trees {
		typ := db.storeTypes[i]
		switch typ {
		case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
			t, err := tree.(*CommitTree).GetImmutable(version)
			if err != nil {
				return nil, fmt.Errorf("failed to create cache multi store for tree %s at version %d: %w", db.treeKeys[i].Name(), version, err)
			}
			mt.trees[i] = t.CacheWrap()
		default:
			mt.trees[i] = tree.CacheWrap()
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
	s := db.trees[index]
	store, ok := s.(storetypes.KVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %v is not KVStore", key))
	}
	return store
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
	s := db.trees[db.treesByKey[key]]
	store, ok := s.(storetypes.CommitKVStore)
	if !ok {
		panic(fmt.Sprintf("store with key %s is not CommitKVStore", key.Name()))
	}
	return store
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
	}

	return nil
}

func (db *CommitMultiTree) loadStore(key storetypes.StoreKey, typ storetypes.StoreType) (storetypes.CommitStore, error) {
	switch typ {
	case storetypes.StoreTypeIAVL, storetypes.StoreTypeDB:
		dir := filepath.Join(db.dir, "stores", key.Name())
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return nil, fmt.Errorf("failed to create store dir %s: %w", dir, err)
			}
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

func LoadDB(path string, opts *Options, logger *slog.Logger) (*CommitMultiTree, error) {
	db := &CommitMultiTree{
		dir:        path,
		opts:       *opts,
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
