package rootmulti

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	tmiavl "github.com/tendermint/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/tmhash"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/cachemulti"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>

	snapshotBufferSize = int(4e6)
	snapshotChunkSize  = uint64(8e6) // Do not change without new snapshot format (must be uniform)
)

var cdc = codec.New()

// Store is composed of many CommitStores. Name contrasts with
// cacheMultiStore which is for cache-wrapping other MultiStores. It implements
// the CommitMultiStore interface.
type Store struct {
	db             dbm.DB
	lastCommitInfo commitInfo
	pruningOpts    types.PruningOptions
	storesParams   map[types.StoreKey]storeParams
	stores         map[types.StoreKey]types.CommitKVStore
	keysByName     map[string]types.StoreKey
	lazyLoading    bool

	traceWriter  io.Writer
	traceContext types.TraceContext

	interBlockCache types.MultiStorePersistentCache
}

var _ types.CommitMultiStore = (*Store)(nil)
var _ types.Queryable = (*Store)(nil)

// NewStore returns a reference to a new Store object with the provided DB. The
// store will be created with a PruneNothing pruning strategy by default. After
// a store is created, KVStores must be mounted and finally LoadLatestVersion or
// LoadVersion must be called.
func NewStore(db dbm.DB) *Store {
	return &Store{
		db:           db,
		pruningOpts:  types.PruneNothing,
		storesParams: make(map[types.StoreKey]storeParams),
		stores:       make(map[types.StoreKey]types.CommitKVStore),
		keysByName:   make(map[string]types.StoreKey),
	}
}

// SetPruning sets the pruning strategy on the root store and all the sub-stores.
// Note, calling SetPruning on the root store prior to LoadVersion or
// LoadLatestVersion performs a no-op as the stores aren't mounted yet.
//
// TODO: Consider removing this API altogether on sub-stores as a pruning
// strategy should only be provided on initialization.
func (rs *Store) SetPruning(pruningOpts types.PruningOptions) {
	rs.pruningOpts = pruningOpts
	for _, substore := range rs.stores {
		substore.SetPruning(pruningOpts)
	}
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
	infos := make(map[string]storeInfo)
	var cInfo commitInfo

	// load old data if we are not version 0
	if ver != 0 {
		var err error
		cInfo, err = getCommitInfo(rs.db, ver)
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
	for key, storeParams := range rs.storesParams {
		// Load it
		store, err := rs.loadCommitStoreFromParams(key, rs.getCommitID(infos, key.Name()), storeParams)
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

	return nil
}

func (rs *Store) getCommitID(infos map[string]storeInfo, name string) types.CommitID {
	info, ok := infos[name]
	if !ok {
		return types.CommitID{}
	}
	return info.Core.CommitID
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

//----------------------------------------
// +CommitStore

// LastCommitID implements Committer/CommitStore.
func (rs *Store) LastCommitID() types.CommitID {
	return rs.lastCommitInfo.CommitID()
}

// Commit implements Committer/CommitStore.
func (rs *Store) Commit() types.CommitID {

	// Commit stores.
	version := rs.lastCommitInfo.Version + 1
	rs.lastCommitInfo = commitStores(version, rs.stores)

	// write CommitInfo to disk only if this version was flushed to disk
	if rs.pruningOpts.FlushVersion(version) {
		flushCommitInfo(rs.db, version, rs.lastCommitInfo)
	}

	// Prepare for next version.
	commitID := types.CommitID{
		Version: version,
		Hash:    rs.lastCommitInfo.Hash(),
	}
	return commitID
}

// CacheWrap implements CacheWrapper/Store/CommitStore.
func (rs *Store) CacheWrap() types.CacheWrap {
	return rs.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (rs *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return rs.CacheWrap()
}

//----------------------------------------
// +MultiStore

// CacheMultiStore cache-wraps the multi-store and returns a CacheMultiStore.
// It implements the MultiStore interface.
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}

	return cachemulti.NewStore(rs.db, stores, rs.keysByName, rs.traceWriter, rs.traceContext)
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

	return cachemulti.NewStore(rs.db, cachedStores, rs.keysByName, rs.traceWriter, rs.traceContext), nil
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
	store := rs.stores[key].(types.KVStore)

	if rs.TracingEnabled() {
		store = tracekv.NewStore(store, rs.traceWriter, rs.traceContext)
	}

	return store
}

// getStoreByName performs a lookup of a StoreKey given a store name typically
// provided in a path. The StoreKey is then used to perform a lookup and return
// a Store. If the Store is wrapped in an inter-block cache, it will be unwrapped
// prior to being returned. If the StoreKey does not exist, nil is returned.
func (rs *Store) getStoreByName(name string) types.Store {
	key := rs.keysByName[name]
	if key == nil {
		return nil
	}

	return rs.GetCommitKVStore(key)
}

//---------------------- Query ------------------

// Query calls substore.Query with the same `req` where `req.Path` is
// modified to remove the substore prefix.
// Ie. `req.Path` here is `/<substore>/<path>`, and trimmed to `/<path>` for the substore.
// TODO: add proof for `multistore -> substore`.
func (rs *Store) Query(req abci.RequestQuery) abci.ResponseQuery {
	path := req.Path
	storeName, subpath, err := parsePath(path)
	if err != nil {
		return sdkerrors.QueryResult(err)
	}

	store := rs.getStoreByName(storeName)
	if store == nil {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "no such store: %s", storeName))
	}

	queryable, ok := store.(types.Queryable)
	if !ok {
		return sdkerrors.QueryResult(sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "store %s (type %T) doesn't support queries", storeName, store))
	}

	// trim the path and make the query
	req.Path = subpath
	res := queryable.Query(req)

	if !req.Prove || !RequireProof(subpath) {
		return res
	}

	if res.Proof == nil || len(res.Proof.Ops) == 0 {
		return sdkerrors.QueryResult(sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "proof is unexpectedly empty; ensure height has not been pruned"))
	}

	// If the request's height is the latest height we've committed, then utilize
	// the store's lastCommitInfo as this commit info may not be flushed to disk.
	// Otherwise, we query for the commit info from disk.
	var commitInfo commitInfo

	if res.Height == rs.lastCommitInfo.Version {
		commitInfo = rs.lastCommitInfo
	} else {
		commitInfo, err = getCommitInfo(rs.db, res.Height)
		if err != nil {
			return sdkerrors.QueryResult(err)
		}
	}

	// Restore origin path and append proof op.
	res.Proof.Ops = append(res.Proof.Ops, NewMultiStoreProofOp(
		[]byte(storeName),
		NewMultiStoreProof(commitInfo.StoreInfos),
	).ProofOp())

	// TODO: handle in another TM v0.26 update PR
	// res.Proof = buildMultiStoreProof(res.Proof, storeName, commitInfo.StoreInfos)
	return res
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

// Snapshot implements Snapshotter.
func (rs *Store) Snapshot(height uint64) (types.Snapshot, error) {

	// Collect stores to snapshot (only IAVL stores are supported)
	type namedStore struct {
		*iavl.Store
		name string
	}
	stores := []namedStore{}
	for key, store := range rs.stores {
		switch store := store.(type) {
		case *iavl.Store:
			stores = append(stores, namedStore{name: key.Name(), Store: store})
		case *transient.Store:
			// Transient stores aren't persisted and shouldn't be snapshotted
			continue
		default:
			return types.Snapshot{}, errors.Errorf("unable to snapshot store %q of type %T",
				key.Name(), store)
		}
	}
	sort.Slice(stores, func(i, j int) bool {
		return strings.Compare(stores[i].name, stores[j].name) == -1
	})

	// Create a channel of snapshot chunk readers, which we'll return to the caller
	ch := make(chan io.ReadCloser)
	chunker, err := newChunkWriter(ch, snapshotChunkSize)
	if err != nil {
		return types.Snapshot{}, err
	}

	// Spawn goroutine to generate snapshot chunks
	go func() {
		// Set up a stream pipeline to serialize snapshot nodes:
		// ExportNode -> delimited Protobuf -> gzip -> buffer -> chunkWriter -> chan io.ReadCloser
		defer chunker.Close()
		bufWriter := bufio.NewWriterSize(chunker, snapshotBufferSize)
		defer func() {
			if err := bufWriter.Flush(); err != nil {
				chunker.CloseWithError(err)
			}
		}()
		gzWriter := gzip.NewWriter(bufWriter)
		defer func() {
			if err := gzWriter.Close(); err != nil {
				chunker.CloseWithError(err)
			}
		}()
		protoWriter := protoio.NewDelimitedWriter(gzWriter)
		defer func() {
			if err := protoWriter.Close(); err != nil {
				chunker.CloseWithError(err)
			}
		}()

		// Export each IAVL store. Stores are serialized as a stream of SnapshotItem Protobuf
		// messages. The first item contains a SnapshotStore with store metadata (i.e. name),
		// and the following messages contain a SnapshotNode (i.e. an ExportNode). Store changes
		// are demarcated by new SnapshotStore items.
		for _, store := range stores {
			exporter, err := store.Export(int64(height))
			if err != nil {
				chunker.CloseWithError(err)
				return
			}
			err = protoWriter.WriteMsg(&sdktypes.SnapshotItem{
				Item: &sdktypes.SnapshotItem_Store{
					Store: &sdktypes.SnapshotStore{
						Name: store.name,
					},
				},
			})
			if err != nil {
				chunker.CloseWithError(err)
				return
			}

			for {
				node, err := exporter.Next()
				if err == tmiavl.ExportDone {
					break
				} else if err != nil {
					chunker.CloseWithError(err)
					return
				}
				err = protoWriter.WriteMsg(&sdktypes.SnapshotItem{
					Item: &sdktypes.SnapshotItem_Node{
						Node: &sdktypes.SnapshotNode{
							Key:     node.Key,
							Value:   node.Value,
							Height:  node.Height,
							Version: node.Version,
						},
					},
				})
				if err != nil {
					chunker.CloseWithError(err)
					return
				}
			}
		}
	}()

	return types.Snapshot{
		Height: height,
		Format: 1,
		Chunks: ch,
	}, nil
}

// Restore implements Snapshotter.
func (rs *Store) Restore(snapshot types.Snapshot) error {
	if snapshot.Format != 1 {
		return errors.Errorf("unsupported snapshot format %v", snapshot.Format)
	}
	if snapshot.Height == 0 {
		return errors.New("cannot restore snapshot at height 0")
	}
	if snapshot.Height > math.MaxInt64 {
		return fmt.Errorf("snapshot height %v cannot exceed %v", snapshot.Height, math.MaxInt64)
	}

	// Set up a restore stream pipeline
	// chan io.ReadCloser -> chunkReader -> gunzip -> delimited Protobuf -> ExportNode
	chunkReader := newChunkReader(snapshot.Chunks)
	defer chunkReader.Close()
	gzReader, err := gzip.NewReader(chunkReader)
	if err != nil {
		return fmt.Errorf("gzip error: %w", err)
	}
	defer gzReader.Close()
	protoReader := protoio.NewDelimitedReader(gzReader, 1e6)
	defer protoReader.Close()

	// Import nodes into stores. The first item is expected to be a SnapshotItem containing
	// a SnapshotStore, telling us which store to import into. The following items will contain
	// SnapshotNode (i.e. ExportNode) until we reach the next SnapshotStore (or EOF).
	var importer *tmiavl.Importer
	for {
		item := &sdktypes.SnapshotItem{}
		err := protoReader.ReadMsg(item)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("invalid protobuf message: %w", err)
		}

		switch item := item.Item.(type) {
		case *sdktypes.SnapshotItem_Store:
			if importer != nil {
				err = importer.Commit()
				if err != nil {
					return fmt.Errorf("IAVL commit failed: %w", err)
				}
			}
			store, ok := rs.getStoreByName(item.Store.Name).(*iavl.Store)
			if !ok || store == nil {
				return fmt.Errorf("cannot import into non-IAVL store %q", item.Store.Name)
			}
			importer, err = store.Import(int64(snapshot.Height))
			if err != nil {
				return fmt.Errorf("import failed: %w", err)
			}

		case *sdktypes.SnapshotItem_Node:
			if importer == nil {
				return fmt.Errorf("received node data before store metadata")
			}
			if item.Node.Height > math.MaxInt8 {
				return fmt.Errorf("node height %v cannot exceed %v", item.Node.Height, math.MaxInt8)
			}
			importer.Add(&tmiavl.ExportNode{
				Key:     item.Node.Key,
				Value:   item.Node.Value,
				Height:  int8(item.Node.Height),
				Version: item.Node.Version,
			})

		default:
			return fmt.Errorf("unknown protobuf message %T", item)
		}
	}

	if importer != nil {
		err := importer.Commit()
		if err != nil {
			return fmt.Errorf("IAVL commit failed: %w", err)
		}
	}

	flushCommitInfo(rs.db, int64(snapshot.Height), rs.buildCommitInfo(int64(snapshot.Height)))
	return rs.LoadLatestVersion()
}

//----------------------------------------
// Note: why do we use key and params.key in different places. Seems like there should be only one key used.
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
		store, err := iavl.LoadStore(db, id, rs.pruningOpts, rs.lazyLoading)
		if err != nil {
			return nil, err
		}

		if rs.interBlockCache != nil {
			// Wrap and get a CommitKVStore with inter-block caching. Note, this should
			// only wrap the primary CommitKVStore, not any store that is already
			// cache-wrapped as that will create unexpected behavior.
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

	default:
		panic(fmt.Sprintf("unrecognized store type %v", params.typ))
	}
}

func (rs *Store) buildCommitInfo(version int64) commitInfo {
	storeInfos := []storeInfo{}
	for key, store := range rs.stores {
		if store.GetStoreType() == types.StoreTypeTransient {
			continue
		}
		storeInfos = append(storeInfos, storeInfo{
			Name: key.Name(),
			Core: storeCore{
				CommitID: store.LastCommitID(),
			},
		})
	}
	return commitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
}

//----------------------------------------
// storeParams

type storeParams struct {
	key types.StoreKey
	db  dbm.DB
	typ types.StoreType
}

//----------------------------------------
// commitInfo

// NOTE: Keep commitInfo a simple immutable struct.
type commitInfo struct {

	// Version
	Version int64

	// Store info for
	StoreInfos []storeInfo
}

// Hash returns the simple merkle root hash of the stores sorted by name.
func (ci commitInfo) Hash() []byte {
	// TODO: cache to ci.hash []byte
	m := make(map[string][]byte, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo.Hash()
	}

	return merkle.SimpleHashFromMap(m)
}

func (ci commitInfo) CommitID() types.CommitID {
	return types.CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

//----------------------------------------
// storeInfo

// storeInfo contains the name and core reference for an
// underlying store.  It is the leaf of the Stores top
// level simple merkle tree.
type storeInfo struct {
	Name string
	Core storeCore
}

type storeCore struct {
	// StoreType StoreType
	CommitID types.CommitID
	// ... maybe add more state
}

// Implements merkle.Hasher.
func (si storeInfo) Hash() []byte {
	// Doesn't write Name, since merkle.SimpleHashFromMap() will
	// include them via the keys.
	bz := si.Core.CommitID.Hash
	hasher := tmhash.New()

	_, err := hasher.Write(bz)
	if err != nil {
		// TODO: Handle with #870
		panic(err)
	}

	return hasher.Sum(nil)
}

//----------------------------------------
// Misc.

func getLatestVersion(db dbm.DB) int64 {
	var latest int64
	latestBytes, err := db.Get([]byte(latestVersionKey))
	if err != nil {
		panic(err)
	} else if latestBytes == nil {
		return 0
	}

	err = cdc.UnmarshalBinaryLengthPrefixed(latestBytes, &latest)
	if err != nil {
		panic(err)
	}

	return latest
}

// Set the latest version.
func setLatestVersion(batch dbm.Batch, version int64) {
	latestBytes, _ := cdc.MarshalBinaryLengthPrefixed(version)
	batch.Set([]byte(latestVersionKey), latestBytes)
}

// Commits each store and returns a new commitInfo.
func commitStores(version int64, storeMap map[types.StoreKey]types.CommitKVStore) commitInfo {
	storeInfos := make([]storeInfo, 0, len(storeMap))

	for key, store := range storeMap {
		commitID := store.Commit()

		if store.GetStoreType() == types.StoreTypeTransient {
			continue
		}

		si := storeInfo{}
		si.Name = key.Name()
		si.Core.CommitID = commitID
		storeInfos = append(storeInfos, si)
	}

	return commitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
}

// Gets commitInfo from disk.
func getCommitInfo(db dbm.DB, ver int64) (commitInfo, error) {

	// Get from DB.
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)
	cInfoBytes, err := db.Get([]byte(cInfoKey))
	if err != nil {
		return commitInfo{}, errors.Wrap(err, "failed to get commit info")
	} else if cInfoBytes == nil {
		return commitInfo{}, errors.New("failed to get commit info: no data")
	}

	var cInfo commitInfo

	err = cdc.UnmarshalBinaryLengthPrefixed(cInfoBytes, &cInfo)
	if err != nil {
		return commitInfo{}, errors.Wrap(err, "failed to get store")
	}

	return cInfo, nil
}

// Set a commitInfo for given version.
func setCommitInfo(batch dbm.Batch, version int64, cInfo commitInfo) {
	cInfoBytes := cdc.MustMarshalBinaryLengthPrefixed(cInfo)
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, version)
	batch.Set([]byte(cInfoKey), cInfoBytes)
}

// flushCommitInfo flushes a commitInfo for given version to the DB. Note, this
// needs to happen atomically.
func flushCommitInfo(db dbm.DB, version int64, cInfo commitInfo) {
	batch := db.NewBatch()
	defer batch.Close()

	setCommitInfo(batch, version, cInfo)
	setLatestVersion(batch, version)
	batch.Write()
}

// chunkWriter reads an input byte stream, splits it into fixed-size chunks, and writes it to a
// sequence of readers via a channel of io.ReadClosers
type chunkWriter struct {
	ch        chan<- io.ReadCloser
	pipe      *io.PipeWriter
	chunkSize uint64
	written   uint64
	closed    bool
}

// newChunkWriter creates a new chunkWriter
func newChunkWriter(ch chan<- io.ReadCloser, chunkSize uint64) (*chunkWriter, error) {
	if chunkSize == 0 {
		return nil, errors.New("chunk size cannot be 0")
	}
	return &chunkWriter{
		ch:        ch,
		chunkSize: chunkSize,
	}, nil
}

// chunk creates a new chunk
func (w *chunkWriter) chunk() error {
	if w.pipe != nil {
		err := w.pipe.Close()
		if err != nil {
			return err
		}
	}
	pr, pw := io.Pipe()
	w.ch <- pr
	w.pipe = pw
	w.written = 0
	return nil
}

// Close implements io.Closer
func (w *chunkWriter) Close() error {
	if !w.closed {
		w.closed = true
		close(w.ch)
		return w.pipe.Close()
	}
	return nil
}

// CloseWithError closes the writer and sends an error to the reader
func (w *chunkWriter) CloseWithError(err error) {
	if !w.closed {
		w.closed = true
		close(w.ch)
		w.pipe.CloseWithError(err)
	}
}

// Write implements io.Writer
func (w *chunkWriter) Write(data []byte) (int, error) {
	if w.closed {
		return 0, errors.New("cannot write to closed chunkWriter")
	}
	nTotal := 0
	for len(data) > 0 {
		if w.pipe == nil || w.written >= w.chunkSize {
			err := w.chunk()
			if err != nil {
				return nTotal, err
			}
		}
		writeSize := w.chunkSize - w.written
		if writeSize > uint64(len(data)) {
			writeSize = uint64(len(data))
		}
		n, err := w.pipe.Write(data[:writeSize])
		w.written += uint64(n)
		nTotal += n
		if err != nil {
			return nTotal, err
		}
		data = data[writeSize:]
	}
	return nTotal, nil
}

// chunkReader combines chunks from a channel of io.ReadClosers into a byte stream
type chunkReader struct {
	ch     <-chan io.ReadCloser
	reader io.ReadCloser
}

// newChunkReader creates a new chunkReader
func newChunkReader(ch <-chan io.ReadCloser) *chunkReader {
	return &chunkReader{ch: ch}
}

// next fetches the next chunk from the channel, or returns io.EOF if there are no more chunks
func (r *chunkReader) next() error {
	reader, ok := <-r.ch
	if !ok {
		return io.EOF
	}
	r.reader = reader
	return nil
}

// Close implements io.ReadCloser
func (r *chunkReader) Close() error {
	if r.reader != nil {
		err := r.reader.Close()
		r.reader = nil
		return err
	}
	return nil
}

// Read implements io.Reader
func (r *chunkReader) Read(p []byte) (int, error) {
	if r.reader == nil {
		err := r.next()
		if err != nil {
			return 0, err
		}
	}
	n, err := r.reader.Read(p)
	if err == io.EOF {
		err = r.reader.Close()
		r.reader = nil
		if err != nil {
			return 0, err
		}
		return r.Read(p)
	}
	return n, err
}
