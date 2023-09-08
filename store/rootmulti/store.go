package rootmulti

import (
	"fmt"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cosmos/cosmos-sdk/pruning"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	iavltree "github.com/cosmos/iavl"
	protoio "github.com/gogo/protobuf/io"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
	dbm "github.com/tendermint/tm-db"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store/cachemulti"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/mem"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vm "github.com/cosmos/cosmos-sdk/x/upgrade/exported"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>
	appVersionKey    = "s/appversion"

	proofsPath = "proofs"
)

type storeParams struct {
	key            types.StoreKey
	db             dbm.DB
	typ            types.StoreType
	initialVersion uint64
}

// Store is composed of many CommitStores. Name contrasts with
// cacheMultiStore which is used for branching other MultiStores. It implements
// the CommitMultiStore interface.
type Store struct {
	db             dbm.DB
	logger         log.Logger
	lastCommitInfo *types.CommitInfo
	mx             sync.RWMutex // mutex to sync access to lastCommitInfo
	pruningManager *pruning.Manager
	iavlCacheSize  int
	storesParams   map[types.StoreKey]storeParams
	stores         map[types.StoreKey]types.CommitKVStore
	keysByName     map[string]types.StoreKey
	lazyLoading    bool
	initialVersion int64

	traceWriter  io.Writer
	traceContext types.TraceContext

	interBlockCache types.MultiStorePersistentCache

	listeners    map[types.StoreKey][]types.WriteListener
	commitHeader tmproto.Header
}

var (
	_ types.CommitMultiStore = (*Store)(nil)
	_ vm.AppVersionManager   = (*Store)(nil)
	_ types.Queryable        = (*Store)(nil)
)

// NewStore returns a reference to a new Store object with the provided DB. The
// store will be created with a PruneNothing pruning strategy by default. After
// a store is created, KVStores must be mounted and finally LoadLatestVersion or
// LoadVersion must be called.
func NewStore(db dbm.DB, logger log.Logger) *Store {
	return &Store{
		db:             db,
		logger:         logger,
		iavlCacheSize:  iavl.DefaultIAVLCacheSize,
		storesParams:   make(map[types.StoreKey]storeParams),
		stores:         make(map[types.StoreKey]types.CommitKVStore),
		keysByName:     make(map[string]types.StoreKey),
		listeners:      make(map[types.StoreKey][]types.WriteListener),
		pruningManager: pruning.NewManager(db, logger),
	}
}

// GetPruning fetches the pruning strategy from the root store.
func (rs *Store) GetPruning() pruningtypes.PruningOptions {
	return rs.pruningManager.GetOptions()
}

// SetPruning sets the pruning strategy on the root store and all the sub-stores.
// Note, calling SetPruning on the root store prior to LoadVersion or
// LoadLatestVersion performs a no-op as the stores aren't mounted yet.
func (rs *Store) SetPruning(pruningOpts pruningtypes.PruningOptions) {
	rs.pruningManager.SetOptions(pruningOpts)
}

// SetSnapshotInterval sets the interval at which the snapshots are taken.
// It is used by the store to determine which heights to retain until after the snapshot is complete.
func (rs *Store) SetSnapshotInterval(snapshotInterval uint64) {
	rs.pruningManager.SetSnapshotInterval(snapshotInterval)
}

func (rs *Store) SetIAVLCacheSize(cacheSize int) {
	rs.iavlCacheSize = cacheSize
}

// SetLazyLoading sets if the iavl store should be loaded lazily or not
func (rs *Store) SetLazyLoading(lazyLoading bool) {
	rs.lazyLoading = lazyLoading
}

// GetStoreType implements Store.
func (rs *Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// MountStoreWithDB implements CommitMultiStore.
func (rs *Store) MountStoreWithDB(key types.StoreKey, typ types.StoreType, db dbm.DB) {
	if key == nil {
		panic("MountIAVLStore() key cannot be nil")
	}
	if _, ok := rs.storesParams[key]; ok {
		panic(fmt.Sprintf("store duplicate store key %v", key))
	}
	if _, ok := rs.keysByName[key.Name()]; ok {
		panic(fmt.Sprintf("store duplicate store key name %v", key))
	}
	rs.storesParams[key] = storeParams{
		key: key,
		typ: typ,
		db:  db,
	}
	rs.keysByName[key.Name()] = key
}

// GetCommitStore returns a mounted CommitStore for a given StoreKey. If the
// store is wrapped in an inter-block cache, it will be unwrapped before returning.
func (rs *Store) GetCommitStore(key types.StoreKey) types.CommitStore {
	return rs.GetCommitKVStore(key)
}

// GetCommitKVStore returns a mounted CommitKVStore for a given StoreKey. If the
// store is wrapped in an inter-block cache, it will be unwrapped before returning.
func (rs *Store) GetCommitKVStore(key types.StoreKey) types.CommitKVStore {
	// If the Store has an inter-block cache, first attempt to lookup and unwrap
	// the underlying CommitKVStore by StoreKey. If it does not exist, fallback to
	// the main mapping of CommitKVStores.
	if rs.interBlockCache != nil {
		if store := rs.interBlockCache.Unwrap(key); store != nil {
			return store
		}
	}

	return rs.stores[key]
}

// StoreKeysByName returns mapping storeNames -> StoreKeys
func (rs *Store) StoreKeysByName() map[string]types.StoreKey {
	return rs.keysByName
}

// GetStores returns mounted stores
func (rs *Store) GetStores() map[types.StoreKey]types.CommitKVStore {
	return rs.stores
}

// LoadLatestVersionAndUpgrade implements CommitMultiStore
func (rs *Store) LoadLatestVersionAndUpgrade(upgrades *types.StoreUpgrades) error {
	ver := getLatestVersion(rs.db)
	return rs.loadVersion(ver, upgrades)
}

// LoadVersionAndUpgrade allows us to rename substores while loading an older version
func (rs *Store) LoadVersionAndUpgrade(ver int64, upgrades *types.StoreUpgrades) error {
	return rs.loadVersion(ver, upgrades)
}

// LoadLatestVersion implements CommitMultiStore.
func (rs *Store) LoadLatestVersion() error {
	ver := getLatestVersion(rs.db)
	return rs.loadVersion(ver, nil)
}

// LoadVersion implements CommitMultiStore.
func (rs *Store) LoadVersion(ver int64) error {
	return rs.loadVersion(ver, nil)
}

func (rs *Store) loadVersion(ver int64, upgrades *types.StoreUpgrades) error {
	infos := make(map[string]types.StoreInfo)

	cInfo := &types.CommitInfo{}

	// load old data if we are not version 0
	if ver != 0 {
		var err error
		cInfo, err = rs.GetCommitInfoFromDB(ver)
		if err != nil {
			return err
		}

		// convert StoreInfos slice to map
		for _, storeInfo := range cInfo.StoreInfos {
			infos[storeInfo.Name] = storeInfo
		}
	}

	// load each Store (note this doesn't panic on unmounted keys now)
	var newStores = make(map[types.StoreKey]types.CommitKVStore)

	storesKeys := make([]types.StoreKey, 0, len(rs.storesParams))

	for key := range rs.storesParams {
		storesKeys = append(storesKeys, key)
	}
	if upgrades != nil {
		// deterministic iteration order for upgrades
		// (as the underlying store may change and
		// upgrades make store changes where the execution order may matter)
		sort.Slice(storesKeys, func(i, j int) bool {
			return storesKeys[i].Name() < storesKeys[j].Name()
		})
	}

	for _, key := range storesKeys {
		storeParams := rs.storesParams[key]
		commitID := rs.getCommitID(infos, key.Name())

		// If it has been added, set the initial version
		if upgrades.IsAdded(key.Name()) {
			storeParams.initialVersion = uint64(ver) + 1
		}

		store, err := rs.loadCommitStoreFromParams(key, commitID, storeParams)
		if err != nil {
			return errors.Wrap(err, "failed to load store")
		}

		newStores[key] = store

		// If it was deleted, remove all data
		if upgrades.IsDeleted(key.Name()) {
			if err := deleteKVStore(store.(types.KVStore)); err != nil {
				return errors.Wrapf(err, "failed to delete store %s", key.Name())
			}
		} else if oldName := upgrades.RenamedFrom(key.Name()); oldName != "" {
			// handle renames specially
			// make an unregistered key to satify loadCommitStore params
			oldKey := types.NewKVStoreKey(oldName)
			oldParams := storeParams
			oldParams.key = oldKey

			// load from the old name
			oldStore, err := rs.loadCommitStoreFromParams(oldKey, rs.getCommitID(infos, oldName), oldParams)
			if err != nil {
				return errors.Wrapf(err, "failed to load old store %s", oldName)
			}

			// move all data
			if err := moveKVStoreData(oldStore.(types.KVStore), store.(types.KVStore)); err != nil {
				return errors.Wrapf(err, "failed to move store %s -> %s", oldName, key.Name())
			}
		}
	}

	rs.lastCommitInfo = cInfo
	rs.stores = newStores

	// load any pruned heights we missed from disk to be pruned on the next run
	if err := rs.pruningManager.LoadPruningHeights(); err != nil {
		return err
	}

	return nil
}

func (rs *Store) getCommitID(infos map[string]types.StoreInfo, name string) types.CommitID {
	info, ok := infos[name]
	if !ok {
		return types.CommitID{}
	}

	return info.CommitId
}

func deleteKVStore(kv types.KVStore) error {
	// Note that we cannot write while iterating, so load all keys here, delete below
	var keys [][]byte
	itr := kv.Iterator(nil, nil)
	for itr.Valid() {
		keys = append(keys, itr.Key())
		itr.Next()
	}
	itr.Close()

	for _, k := range keys {
		kv.Delete(k)
	}
	return nil
}

// we simulate move by a copy and delete
func moveKVStoreData(oldDB types.KVStore, newDB types.KVStore) error {
	// we read from one and write to another
	itr := oldDB.Iterator(nil, nil)
	for itr.Valid() {
		newDB.Set(itr.Key(), itr.Value())
		itr.Next()
	}
	itr.Close()

	// then delete the old store
	return deleteKVStore(oldDB)
}

// PruneSnapshotHeight prunes the given height according to the prune strategy.
// If PruneNothing, this is a no-op.
// If other strategy, this height is persisted until it is
// less than <current height> - KeepRecent and <current height> % Interval == 0
func (rs *Store) PruneSnapshotHeight(height int64) {
	rs.pruningManager.HandleHeightSnapshot(height)
}

// SetInterBlockCache sets the Store's internal inter-block (persistent) cache.
// When this is defined, all CommitKVStores will be wrapped with their respective
// inter-block cache.
func (rs *Store) SetInterBlockCache(c types.MultiStorePersistentCache) {
	rs.interBlockCache = c
}

// SetTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (rs *Store) SetTracer(w io.Writer) types.MultiStore {
	rs.traceWriter = w
	return rs
}

// SetTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (rs *Store) SetTracingContext(tc types.TraceContext) types.MultiStore {
	if rs.traceContext != nil {
		for k, v := range tc {
			rs.traceContext[k] = v
		}
	} else {
		rs.traceContext = tc
	}

	return rs
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (rs *Store) TracingEnabled() bool {
	return rs.traceWriter != nil
}

// AddListeners adds listeners for a specific KVStore
func (rs *Store) AddListeners(key types.StoreKey, listeners []types.WriteListener) {
	if ls, ok := rs.listeners[key]; ok {
		rs.listeners[key] = append(ls, listeners...)
	} else {
		rs.listeners[key] = listeners
	}
}

// ListeningEnabled returns if listening is enabled for a specific KVStore
func (rs *Store) ListeningEnabled(key types.StoreKey) bool {
	if ls, ok := rs.listeners[key]; ok {
		return len(ls) != 0
	}
	return false
}

// LastCommitID implements Committer/CommitStore.
func (rs *Store) LastCommitID() types.CommitID {
	rs.mx.RLock()
	defer rs.mx.RUnlock()
	if rs.lastCommitInfo == nil {
		return types.CommitID{
			Version: getLatestVersion(rs.db),
		}
	}

	return rs.lastCommitInfo.CommitID()
}

// Commit implements Committer/CommitStore.
func (rs *Store) Commit() types.CommitID {
	var previousHeight, version int64
	if rs.lastCommitInfo.GetVersion() == 0 && rs.initialVersion > 1 {
		// This case means that no commit has been made in the store, we
		// start from initialVersion.
		version = rs.initialVersion

	} else {
		// This case can means two things:
		// - either there was already a previous commit in the store, in which
		// case we increment the version from there,
		// - or there was no previous commit, and initial version was not set,
		// in which case we start at version 1.
		previousHeight = rs.lastCommitInfo.GetVersion()
		version = previousHeight + 1
	}

	newCommitInfo := rs.commitStores(version, rs.stores)
	newCommitInfo.Timestamp = rs.commitHeader.Time
	rs.updateLatestCommitInfo(newCommitInfo, version)

	err := rs.handlePruning(version)
	if err != nil {
		panic(err)
	}

	rs.mx.RLock()
	defer rs.mx.RUnlock()
	hash, keys := rs.lastCommitInfo.Hash()
	rs.logger.Debug("calculated commit hash", "height", version, "commit_hash", fmt.Sprintf("%X", hash), "keys", keys)

	return types.CommitID{
		Version: version,
		Hash:    hash,
	}
}

// SetCommitHeader sets the commit block header of the store.
func (rs *Store) SetCommitHeader(h tmproto.Header) {
	rs.commitHeader = h
}

// CacheWrap implements CacheWrapper/Store/CommitStore.
func (rs *Store) CacheWrap() types.CacheWrap {
	return rs.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (rs *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return rs.CacheWrap()
}

// CacheWrapWithListeners implements the CacheWrapper interface.
func (rs *Store) CacheWrapWithListeners(_ types.StoreKey, _ []types.WriteListener) types.CacheWrap {
	return rs.CacheWrap()
}

// CacheMultiStore creates ephemeral branch of the multi-store and returns a CacheMultiStore.
// It implements the MultiStore interface.
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}
	return cachemulti.NewStore(rs.db, stores, rs.keysByName, rs.traceWriter, rs.traceContext, rs.listeners)
}

// CacheMultiStoreWithVersion is analogous to CacheMultiStore except that it
// attempts to load stores at a given version (height). An error is returned if
// any store cannot be loaded. This should only be used for querying and
// iterating at past heights.
func (rs *Store) CacheMultiStoreWithVersion(version int64) (types.CacheMultiStore, error) {
	cachedStores := make(map[types.StoreKey]types.CacheWrapper)
	for key, store := range rs.stores {
		switch store.GetStoreType() {
		case types.StoreTypeIAVL:
			// If the store is wrapped with an inter-block cache, we must first unwrap
			// it to get the underlying IAVL store.
			store = rs.GetCommitKVStore(key)

			// Attempt to lazy-load an already saved IAVL store version. If the
			// version does not exist or is pruned, an error should be returned.
			iavlStore, err := store.(*iavl.Store).GetImmutable(version)
			if err != nil {
				return nil, err
			}

			cachedStores[key] = iavlStore

		default:
			cachedStores[key] = store
		}
	}

	return cachemulti.NewStore(rs.db, cachedStores, rs.keysByName, rs.traceWriter, rs.traceContext, rs.listeners), nil
}

// GetStore returns a mounted Store for a given StoreKey. If the StoreKey does
// not exist, it will panic. If the Store is wrapped in an inter-block cache, it
// will be unwrapped prior to being returned.
//
// TODO: This isn't used directly upstream. Consider returning the Store as-is
// instead of unwrapping.
func (rs *Store) GetStore(key types.StoreKey) types.Store {
	store := rs.GetCommitKVStore(key)
	if store == nil {
		panic(fmt.Sprintf("store does not exist for key: %s", key.Name()))
	}

	return store
}

// GetKVStore returns a mounted KVStore for a given StoreKey. If tracing is
// enabled on the KVStore, a wrapped TraceKVStore will be returned with the root
// store's tracer, otherwise, the original KVStore will be returned.
//
// NOTE: The returned KVStore may be wrapped in an inter-block cache if it is
// set on the root store.
func (rs *Store) GetKVStore(key types.StoreKey) types.KVStore {
	s := rs.stores[key]
	if s == nil {
		panic(fmt.Sprintf("store does not exist for key: %s", key.Name()))
	}
	store := s.(types.KVStore)

	if rs.TracingEnabled() {
		store = tracekv.NewStore(store, rs.traceWriter, rs.traceContext)
	}
	if rs.ListeningEnabled(key) {
		store = listenkv.NewStore(store, key, rs.listeners[key])
	}

	return store
}

func (rs *Store) handlePruning(version int64) error {
	rs.pruningManager.HandleHeight(version - 1) // we should never prune the current version.
	if !rs.pruningManager.ShouldPruneAtHeight(version) {
		return nil
	}
	rs.logger.Info("prune start", "height", version)
	defer rs.logger.Info("prune end", "height", version)
	return rs.pruneStores()
}

func (rs *Store) pruneStores() error {
	pruningHeights, err := rs.pruningManager.GetFlushAndResetPruningHeights()
	if err != nil {
		return err
	}

	if len(pruningHeights) == 0 {
		rs.logger.Debug("pruning skipped; no heights to prune")
		return nil
	}

	rs.logger.Debug("pruning heights", "heights", pruningHeights)

	for key, store := range rs.stores {
		// If the store is wrapped with an inter-block cache, we must first unwrap
		// it to get the underlying IAVL store.
		if store.GetStoreType() != types.StoreTypeIAVL {
			continue
		}

		store = rs.GetCommitKVStore(key)

		err := store.(*iavl.Store).DeleteVersions(pruningHeights...)
		if err == nil {
			continue
		}

		if errCause := errors.Cause(err); errCause != nil && errCause != iavltree.ErrVersionDoesNotExist {
			return err
		}
	}
	return nil
}

// getStoreByName performs a lookup of a StoreKey given a store name typically
// provided in a path. The StoreKey is then used to perform a lookup and return
// a Store. If the Store is wrapped in an inter-block cache, it will be unwrapped
// prior to being returned. If the StoreKey does not exist, nil is returned.
func (rs *Store) GetStoreByName(name string) types.Store {
	key := rs.keysByName[name]
	if key == nil {
		return nil
	}

	return rs.GetCommitKVStore(key)
}

// Query calls substore.Query with the same `req` where `req.Path` is
// modified to remove the substore prefix.
// Ie. `req.Path` here is `/<substore>/<path>`, and trimmed to `/<path>` for the substore.
// Special case: if `req.Path` is `/proofs`, the commit hash is included
// as response value. In addition, proofs of every store are appended to the response for
// the requested height
func (rs *Store) Query(req abci.RequestQuery) abci.ResponseQuery {
	path := req.Path
	firstPath, subpath, err := parsePath(path)
	if err != nil {
		return sdkerrors.QueryResult(err)
	}

	if firstPath == proofsPath {
		return rs.doProofsQuery(req)
	}

	store := rs.GetStoreByName(firstPath)
	if store == nil {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "no such store: %s", firstPath))
	}

	queryable, ok := store.(types.Queryable)
	if !ok {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "store %s (type %T) doesn't support queries", firstPath, store))
	}

	// trim the path and make the query
	req.Path = subpath
	res := queryable.Query(req)

	if !req.Prove || !RequireProof(subpath) {
		return res
	}

	if res.ProofOps == nil || len(res.ProofOps.Ops) == 0 {
		return sdkerrors.QueryResult(sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proof is unexpectedly empty; ensure height has not been pruned"))
	}

	commitInfo, err := rs.GetCommitInfoFromDB(res.Height)
	if err != nil {
		return sdkerrors.QueryResult(err)
	}

	// Restore origin path and append proof op.
	res.ProofOps.Ops = append(res.ProofOps.Ops, commitInfo.ProofOp(firstPath))

	return res
}

// SetInitialVersion sets the initial version of the IAVL tree. It is used when
// starting a new chain at an arbitrary height.
func (rs *Store) SetInitialVersion(version int64) error {
	rs.initialVersion = version

	// Loop through all the stores, if it's an IAVL store, then set initial
	// version on it.
	for key, store := range rs.stores {
		if store.GetStoreType() == types.StoreTypeIAVL {
			// If the store is wrapped with an inter-block cache, we must first unwrap
			// it to get the underlying IAVL store.
			store = rs.GetCommitKVStore(key)
			store.(*iavl.Store).SetInitialVersion(version)
		}
	}

	return nil
}

// parsePath expects a format like /<storeName>[/<subpath>]
// Must start with /, subpath may be empty
// Returns error if it doesn't start with /
func parsePath(path string) (storeName string, subpath string, err error) {
	if !strings.HasPrefix(path, "/") {
		return storeName, subpath, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid path: %s", path)
	}

	paths := strings.SplitN(path[1:], "/", 2)
	storeName = paths[0]

	if len(paths) == 2 {
		subpath = "/" + paths[1]
	}

	return storeName, subpath, nil
}

//---------------------- Snapshotting ------------------

// Snapshot implements snapshottypes.Snapshotter. The snapshot output for a given format must be
// identical across nodes such that chunks from different sources fit together. If the output for a
// given format changes (at the byte level), the snapshot format must be bumped - see
// TestMultistoreSnapshot_Checksum test.
func (rs *Store) Snapshot(height uint64, protoWriter protoio.Writer) error {
	if height == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrLogic, "cannot snapshot height 0")
	}
	if height > uint64(getLatestVersion(rs.db)) {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot snapshot future height %v", height)
	}

	// Collect stores to snapshot (only IAVL stores are supported)
	type namedStore struct {
		*iavl.Store
		name string
	}
	stores := []namedStore{}
	for key := range rs.stores {
		switch store := rs.GetCommitKVStore(key).(type) {
		case *iavl.Store:
			stores = append(stores, namedStore{name: key.Name(), Store: store})
		case *transient.Store, *mem.Store:
			// Non-persisted stores shouldn't be snapshotted
			continue
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrLogic,
				"don't know how to snapshot store %q of type %T", key.Name(), store)
		}
	}
	sort.Slice(stores, func(i, j int) bool {
		return strings.Compare(stores[i].name, stores[j].name) == -1
	})

	appVersion, err := rs.GetAppVersion()
	if err != nil {
		return err
	}

	err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
		Item: &snapshottypes.SnapshotItem_AppVersion{
			AppVersion: &snapshottypes.SnapshotAppVersion{
				Version: appVersion,
			},
		},
	})
	if err != nil {
		return err
	}

	// Export each IAVL store. Stores are serialized as a stream of SnapshotItem Protobuf
	// messages. The first item contains a SnapshotStore with store metadata (i.e. name),
	// and the following messages contain a SnapshotNode (i.e. an ExportNode). Store changes
	// are demarcated by new SnapshotStore items.
	for _, store := range stores {
		exporter, err := store.Export(int64(height))
		if err != nil {
			return err
		}
		defer exporter.Close()
		err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
			Item: &snapshottypes.SnapshotItem_Store{
				Store: &snapshottypes.SnapshotStoreItem{
					Name: store.name,
				},
			},
		})
		if err != nil {
			return err
		}

		for {
			node, err := exporter.Next()
			if err == iavltree.ExportDone {
				break
			} else if err != nil {
				return err
			}
			err = protoWriter.WriteMsg(&snapshottypes.SnapshotItem{
				Item: &snapshottypes.SnapshotItem_IAVL{
					IAVL: &snapshottypes.SnapshotIAVLItem{
						Key:     node.Key,
						Value:   node.Value,
						Height:  int32(node.Height),
						Version: node.Version,
					},
				},
			})
			if err != nil {
				return err
			}
		}
		exporter.Close()
	}

	return nil
}

// Restore implements snapshottypes.Snapshotter.
// returns next snapshot item and error.
func (rs *Store) Restore(
	height uint64, format uint32, protoReader protoio.Reader,
) (snapshottypes.SnapshotItem, error) {
	// Import nodes into stores. The first item is expected to be a SnapshotItem containing
	// a SnapshotStoreItem, telling us which store to import into. The following items will contain
	// SnapshotNodeItem (i.e. ExportNode) until we reach the next SnapshotStoreItem or EOF.
	var importer *iavltree.Importer
	var snapshotItem snapshottypes.SnapshotItem
loop:
	for {
		snapshotItem = snapshottypes.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if err == io.EOF {
			break
		} else if err != nil {
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "invalid protobuf message")
		}

		switch item := snapshotItem.Item.(type) {
		case *snapshottypes.SnapshotItem_Store:
			if importer != nil {
				err = importer.Commit()
				if err != nil {
					return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "IAVL commit failed")
				}
				importer.Close()
			}
			store, ok := rs.GetStoreByName(item.Store.Name).(*iavl.Store)
			if !ok || store == nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot import into non-IAVL store %q", item.Store.Name)
			}
			importer, err = store.Import(int64(height))
			if err != nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "import failed")
			}
			defer importer.Close()

		case *snapshottypes.SnapshotItem_IAVL:
			if importer == nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(sdkerrors.ErrLogic, "received IAVL node item before store item")
			}
			if item.IAVL.Height > math.MaxInt8 {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrapf(sdkerrors.ErrLogic, "node height %v cannot exceed %v",
					item.IAVL.Height, math.MaxInt8)
			}
			node := &iavltree.ExportNode{
				Key:     item.IAVL.Key,
				Value:   item.IAVL.Value,
				Height:  int8(item.IAVL.Height),
				Version: item.IAVL.Version,
			}
			// Protobuf does not differentiate between []byte{} as nil, but fortunately IAVL does
			// not allow nil keys nor nil values for leaf nodes, so we can always set them to empty.
			if node.Key == nil {
				node.Key = []byte{}
			}
			if node.Height == 0 && node.Value == nil {
				node.Value = []byte{}
			}
			err := importer.Add(node)
			if err != nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "IAVL node import failed")
			}

		case *snapshottypes.SnapshotItem_AppVersion:
			err := rs.SetAppVersion(uint64(item.AppVersion.Version))
			if err != nil {
				return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "IAVL node import failed - error saving app version from received snapshot")
			}

		default:
			break loop
		}
	}

	if importer != nil {
		err := importer.Commit()
		if err != nil {
			return snapshottypes.SnapshotItem{}, sdkerrors.Wrap(err, "IAVL commit failed")
		}
		importer.Close()
	}

	rs.flushLastCommitInfo(rs.buildCommitInfo(int64(height)))
	return snapshotItem, rs.LoadLatestVersion()
}

func (rs *Store) loadCommitStoreFromParams(key types.StoreKey, id types.CommitID, params storeParams) (types.CommitKVStore, error) {
	var db dbm.DB

	if params.db != nil {
		db = dbm.NewPrefixDB(params.db, []byte("s/_/"))
	} else {
		prefix := "s/k:" + params.key.Name() + "/"
		db = dbm.NewPrefixDB(rs.db, []byte(prefix))
	}

	switch params.typ {
	case types.StoreTypeMulti:
		panic("recursive MultiStores not yet supported")

	case types.StoreTypeIAVL:
		var store types.CommitKVStore
		var err error

		if params.initialVersion == 0 {
			store, err = iavl.LoadStore(db, rs.logger, key, id, rs.lazyLoading, rs.iavlCacheSize)
		} else {
			store, err = iavl.LoadStoreWithInitialVersion(db, rs.logger, key, id, rs.lazyLoading, params.initialVersion, rs.iavlCacheSize)
		}

		if err != nil {
			return nil, err
		}

		if rs.interBlockCache != nil {
			// Wrap and get a CommitKVStore with inter-block caching. Note, this should
			// only wrap the primary CommitKVStore, not any store that is already
			// branched as that will create unexpected behavior.
			store = rs.interBlockCache.GetStoreCache(key, store)
		}

		return store, err

	case types.StoreTypeDB:
		return commitDBStoreAdapter{Store: dbadapter.Store{DB: db}}, nil

	case types.StoreTypeTransient:
		_, ok := key.(*types.TransientStoreKey)
		if !ok {
			return nil, fmt.Errorf("invalid StoreKey for StoreTypeTransient: %s", key.String())
		}

		return transient.NewStore(), nil

	case types.StoreTypeMemory:
		if _, ok := key.(*types.MemoryStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a MemoryStoreKey; got: %s", key.String())
		}

		return mem.NewStore(), nil

	default:
		panic(fmt.Sprintf("unrecognized store type %v", params.typ))
	}
}

func (rs *Store) buildCommitInfo(version int64) *types.CommitInfo {
	storeInfos := []types.StoreInfo{}
	for key, store := range rs.stores {
		if store.GetStoreType() == types.StoreTypeTransient {
			continue
		}
		storeInfos = append(storeInfos, types.StoreInfo{
			Name:     key.Name(),
			CommitId: store.LastCommitID(),
		})
	}
	return &types.CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
}

// Commits each store and returns a new commitInfo.
func (rs *Store) commitStores(version int64, storeMap map[types.StoreKey]types.CommitKVStore) *types.CommitInfo {
	storeInfos := make([]types.StoreInfo, 0, len(storeMap))

	for key, store := range storeMap {
		commitID := store.Commit()
		rs.logger.Debug("committed KVStore", "height", commitID.Version, "key", key.Name(), "commit_store_hash", fmt.Sprintf("%X", commitID.Hash))

		if store.GetStoreType() == types.StoreTypeTransient {
			continue
		}

		si := types.StoreInfo{}
		si.Name = key.Name()
		si.CommitId = commitID
		storeInfos = append(storeInfos, si)
	}

	return &types.CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
}

func (rs *Store) updateLatestCommitInfo(newCommitInfo *types.CommitInfo, version int64) {
	rs.mx.Lock()
	defer rs.mx.Unlock()
	rs.lastCommitInfo = newCommitInfo
	rs.flushLastCommitInfo(newCommitInfo)
}

func (rs *Store) flushLastCommitInfo(cInfo *types.CommitInfo) {
	batch := rs.db.NewBatch()
	defer batch.Close()

	setCommitInfo(batch, cInfo)
	setLatestVersion(batch, cInfo.Version)

	if err := batch.WriteSync(); err != nil {
		panic(fmt.Errorf("error on batch write %w", err))
	}
}

// Gets commitInfo from disk.
func (rs *Store) GetCommitInfoFromDB(ver int64) (*types.CommitInfo, error) {
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)

	bz, err := rs.db.Get([]byte(cInfoKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get commit info")
	} else if bz == nil {
		return nil, errors.New("no commit info found")
	}

	cInfo := &types.CommitInfo{}
	if err = cInfo.Unmarshal(bz); err != nil {
		return nil, errors.Wrap(err, "failed unmarshal commit info")
	}

	return cInfo, nil
}

func (rs *Store) SetAppVersion(appVersion uint64) error {
	if err := rs.db.SetSync([]byte(appVersionKey), []byte(strconv.Itoa(int(appVersion)))); err != nil {
		return fmt.Errorf("error on write %w", err)
	}
	return nil
}

func (rs *Store) GetAppVersion() (uint64, error) {
	bz, err := rs.db.Get([]byte(appVersionKey))
	if err != nil {
		return 0, errors.Wrap(err, "failed to get app version")
	} else if bz == nil {
		return 0, nil
	}

	appVersion, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		return 0, err
	}

	return appVersion, nil
}

func (rs *Store) doProofsQuery(req abci.RequestQuery) abci.ResponseQuery {
	commitInfo, err := rs.GetCommitInfoFromDB(req.Height)
	if err != nil {
		return sdkerrors.QueryResult(err)
	}
	res := abci.ResponseQuery{
		Height:   req.Height,
		Key:      []byte(proofsPath),
		Value:    commitInfo.CommitID().Hash,
		ProofOps: &crypto.ProofOps{Ops: make([]crypto.ProofOp, 0, len(commitInfo.StoreInfos))},
	}

	for _, storeInfo := range commitInfo.StoreInfos {
		res.ProofOps.Ops = append(res.ProofOps.Ops, crypto.ProofOp{Key: []byte(storeInfo.Name), Data: storeInfo.CommitId.Hash})
	}
	return res
}

func getLatestVersion(db dbm.DB) int64 {
	bz, err := db.Get([]byte(latestVersionKey))
	if err != nil {
		panic(err)
	} else if bz == nil {
		return 0
	}

	var latestVersion int64

	if err := gogotypes.StdInt64Unmarshal(&latestVersion, bz); err != nil {
		panic(err)
	}

	return latestVersion
}

func setCommitInfo(batch dbm.Batch, cInfo *types.CommitInfo) {
	bz, err := cInfo.Marshal()
	if err != nil {
		panic(err)
	}

	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, cInfo.Version)
	batch.Set([]byte(cInfoKey), bz)
}

func setLatestVersion(batch dbm.Batch, version int64) {
	bz, err := gogotypes.StdInt64Marshal(version)
	if err != nil {
		panic(err)
	}

	batch.Set([]byte(latestVersionKey), bz)
}
