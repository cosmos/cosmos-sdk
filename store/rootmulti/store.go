package rootmulti

import (
	"fmt"
	"io"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/tmhash"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store/cachemulti"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/cosmos/cosmos-sdk/store/errors"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	"github.com/cosmos/cosmos-sdk/store/transient"
	"github.com/cosmos/cosmos-sdk/store/types"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>
)

// Store is composed of many CommitStores. Name contrasts with
// cacheMultiStore which is for cache-wrapping other MultiStores. It implements
// the CommitMultiStore interface.
type Store struct {
	db           dbm.DB
	lastCommitID types.CommitID
	pruningOpts  types.PruningOptions
	storesParams map[types.StoreKey]storeParams
	stores       map[types.StoreKey]types.CommitStore
	keysByName   map[string]types.StoreKey

	traceWriter  io.Writer
	traceContext types.TraceContext
}

var _ types.CommitMultiStore = (*Store)(nil)
var _ types.Queryable = (*Store)(nil)

// nolint
func NewStore(db dbm.DB) *Store {
	return &Store{
		db:           db,
		storesParams: make(map[types.StoreKey]storeParams),
		stores:       make(map[types.StoreKey]types.CommitStore),
		keysByName:   make(map[string]types.StoreKey),
	}
}

// Implements CommitMultiStore
func (rs *Store) SetPruning(pruningOpts types.PruningOptions) {
	rs.pruningOpts = pruningOpts
	for _, substore := range rs.stores {
		substore.SetPruning(pruningOpts)
	}
}

// Implements Store.
func (rs *Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Implements CommitMultiStore.
func (rs *Store) MountStoreWithDB(key types.StoreKey, typ types.StoreType, db dbm.DB) {
	if key == nil {
		panic("MountIAVLStore() key cannot be nil")
	}
	if _, ok := rs.storesParams[key]; ok {
		panic(fmt.Sprintf("Store duplicate store key %v", key))
	}
	if _, ok := rs.keysByName[key.Name()]; ok {
		panic(fmt.Sprintf("Store duplicate store key name %v", key))
	}
	rs.storesParams[key] = storeParams{
		key: key,
		typ: typ,
		db:  db,
	}
	rs.keysByName[key.Name()] = key
}

// Implements CommitMultiStore.
func (rs *Store) GetCommitStore(key types.StoreKey) types.CommitStore {
	return rs.stores[key]
}

// Implements CommitMultiStore.
func (rs *Store) GetCommitKVStore(key types.StoreKey) types.CommitKVStore {
	return rs.stores[key].(types.CommitKVStore)
}

// Implements CommitMultiStore.
func (rs *Store) LoadLatestVersion() error {
	ver := getLatestVersion(rs.db)
	return rs.LoadVersion(ver)
}

// Implements CommitMultiStore.
func (rs *Store) LoadVersion(ver int64) error {

	// Special logic for version 0
	if ver == 0 {
		for key, storeParams := range rs.storesParams {
			id := types.CommitID{}
			store, err := rs.loadCommitStoreFromParams(key, id, storeParams)
			if err != nil {
				return fmt.Errorf("failed to load Store: %v", err)
			}
			rs.stores[key] = store
		}

		rs.lastCommitID = types.CommitID{}
		return nil
	}
	// Otherwise, version is 1 or greater

	// Get commitInfo
	cInfo, err := getCommitInfo(rs.db, ver)
	if err != nil {
		return err
	}

	// Convert StoreInfos slice to map
	infos := make(map[types.StoreKey]storeInfo)
	for _, storeInfo := range cInfo.StoreInfos {
		infos[rs.nameToKey(storeInfo.Name)] = storeInfo
	}

	// Load each Store
	var newStores = make(map[types.StoreKey]types.CommitStore)
	for key, storeParams := range rs.storesParams {
		var id types.CommitID
		info, ok := infos[key]
		if ok {
			id = info.Core.CommitID
		}

		store, err := rs.loadCommitStoreFromParams(key, id, storeParams)
		if err != nil {
			return fmt.Errorf("failed to load Store: %v", err)
		}
		newStores[key] = store
	}

	// Success.
	rs.lastCommitID = cInfo.CommitID()
	rs.stores = newStores
	return nil
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

// Implements Committer/CommitStore.
func (rs *Store) LastCommitID() types.CommitID {
	return rs.lastCommitID
}

// Implements Committer/CommitStore.
func (rs *Store) Commit() types.CommitID {

	// Commit stores.
	version := rs.lastCommitID.Version + 1
	commitInfo := commitStores(version, rs.stores)

	// Need to update atomically.
	batch := rs.db.NewBatch()
	setCommitInfo(batch, version, commitInfo)
	setLatestVersion(batch, version)
	batch.Write()

	// Prepare for next version.
	commitID := types.CommitID{
		Version: version,
		Hash:    commitInfo.Hash(),
	}
	rs.lastCommitID = commitID
	return commitID
}

// Implements CacheWrapper/Store/CommitStore.
func (rs *Store) CacheWrap() types.CacheWrap {
	return rs.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (rs *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return rs.CacheWrap()
}

//----------------------------------------
// +MultiStore

// Implements MultiStore.
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}
	return cachemulti.NewStore(rs.db, stores, rs.keysByName, rs.traceWriter, rs.traceContext)
}

// Implements MultiStore.
// If the store does not exist, panics.
func (rs *Store) GetStore(key types.StoreKey) types.Store {
	store := rs.stores[key]
	if store == nil {
		panic("Could not load store " + key.String())
	}
	return store
}

// GetKVStore implements the MultiStore interface. If tracing is enabled on the
// Store, a wrapped TraceKVStore will be returned with the given
// tracer, otherwise, the original KVStore will be returned.
// If the store does not exist, panics.
func (rs *Store) GetKVStore(key types.StoreKey) types.KVStore {
	store := rs.stores[key].(types.KVStore)

	if rs.TracingEnabled() {
		store = tracekv.NewStore(store, rs.traceWriter, rs.traceContext)
	}

	return store
}

// Implements MultiStore

// getStoreByName will first convert the original name to
// a special key, before looking up the CommitStore.
// This is not exposed to the extensions (which will need the
// StoreKey), but is useful in main, and particularly app.Query,
// in order to convert human strings into CommitStores.
func (rs *Store) getStoreByName(name string) types.Store {
	key := rs.keysByName[name]
	if key == nil {
		return nil
	}
	return rs.stores[key]
}

//---------------------- Query ------------------

// Query calls substore.Query with the same `req` where `req.Path` is
// modified to remove the substore prefix.
// Ie. `req.Path` here is `/<substore>/<path>`, and trimmed to `/<path>` for the substore.
// TODO: add proof for `multistore -> substore`.
func (rs *Store) Query(req abci.RequestQuery) abci.ResponseQuery {
	// Query just routes this to a substore.
	path := req.Path
	storeName, subpath, err := parsePath(path)
	if err != nil {
		return err.QueryResult()
	}

	store := rs.getStoreByName(storeName)
	if store == nil {
		msg := fmt.Sprintf("no such store: %s", storeName)
		return errors.ErrUnknownRequest(msg).QueryResult()
	}
	queryable, ok := store.(types.Queryable)
	if !ok {
		msg := fmt.Sprintf("store %s doesn't support queries", storeName)
		return errors.ErrUnknownRequest(msg).QueryResult()
	}

	// trim the path and make the query
	req.Path = subpath
	res := queryable.Query(req)

	if !req.Prove || !RequireProof(subpath) {
		return res
	}

	if res.Proof == nil || len(res.Proof.Ops) == 0 {
		return errors.ErrInternal("substore proof was nil/empty when it should never be").QueryResult()
	}

	commitInfo, errMsg := getCommitInfo(rs.db, res.Height)
	if errMsg != nil {
		return errors.ErrInternal(errMsg.Error()).QueryResult()
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
func parsePath(path string) (storeName string, subpath string, err errors.Error) {
	if !strings.HasPrefix(path, "/") {
		err = errors.ErrUnknownRequest(fmt.Sprintf("invalid path: %s", path))
		return
	}

	paths := strings.SplitN(path[1:], "/", 2)
	storeName = paths[0]

	if len(paths) == 2 {
		subpath = "/" + paths[1]
	}

	return
}

//----------------------------------------

func (rs *Store) loadCommitStoreFromParams(key types.StoreKey, id types.CommitID, params storeParams) (store types.CommitStore, err error) {
	var db dbm.DB
	if params.db != nil {
		db = dbm.NewPrefixDB(params.db, []byte("s/_/"))
	} else {
		db = dbm.NewPrefixDB(rs.db, []byte("s/k:"+params.key.Name()+"/"))
	}
	switch params.typ {
	case types.StoreTypeMulti:
		panic("recursive MultiStores not yet supported")
		// TODO: id?
		// return NewCommitMultiStore(db, id)
	case types.StoreTypeIAVL:
		store, err = iavl.LoadStore(db, id, rs.pruningOpts)
		return
	case types.StoreTypeDB:
		store = commitDBStoreAdapter{dbadapter.Store{db}}
		return
	case types.StoreTypeTransient:
		_, ok := key.(*types.TransientStoreKey)
		if !ok {
			err = fmt.Errorf("invalid StoreKey for StoreTypeTransient: %s", key.String())
			return
		}
		store = transient.NewStore()
		return
	default:
		panic(fmt.Sprintf("unrecognized store type %v", params.typ))
	}
}

func (rs *Store) nameToKey(name string) types.StoreKey {
	for key := range rs.storesParams {
		if key.Name() == name {
			return key
		}
	}
	panic("Unknown name " + name)
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
	bz, _ := cdc.MarshalBinaryLengthPrefixed(si.Core)
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
	latestBytes := db.Get([]byte(latestVersionKey))
	if latestBytes == nil {
		return 0
	}

	err := cdc.UnmarshalBinaryLengthPrefixed(latestBytes, &latest)
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
func commitStores(version int64, storeMap map[types.StoreKey]types.CommitStore) commitInfo {
	storeInfos := make([]storeInfo, 0, len(storeMap))

	for key, store := range storeMap {
		// Commit
		commitID := store.Commit()

		if store.GetStoreType() == types.StoreTypeTransient {
			continue
		}

		// Record CommitID
		si := storeInfo{}
		si.Name = key.Name()
		si.Core.CommitID = commitID
		// si.Core.StoreType = store.GetStoreType()
		storeInfos = append(storeInfos, si)
	}

	ci := commitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
	return ci
}

// Gets commitInfo from disk.
func getCommitInfo(db dbm.DB, ver int64) (commitInfo, error) {

	// Get from DB.
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)
	cInfoBytes := db.Get([]byte(cInfoKey))
	if cInfoBytes == nil {
		return commitInfo{}, fmt.Errorf("failed to get Store: no data")
	}

	var cInfo commitInfo

	err := cdc.UnmarshalBinaryLengthPrefixed(cInfoBytes, &cInfo)
	if err != nil {
		return commitInfo{}, fmt.Errorf("failed to get Store: %v", err)
	}

	return cInfo, nil
}

// Set a commitInfo for given version.
func setCommitInfo(batch dbm.Batch, version int64, cInfo commitInfo) {
	cInfoBytes := cdc.MustMarshalBinaryLengthPrefixed(cInfo)
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, version)
	batch.Set([]byte(cInfoKey), cInfoBytes)
}
