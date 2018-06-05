package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/ripemd160"

	abci "github.com/tendermint/abci/types"
	libmerkle "github.com/tendermint/go-crypto/merkle"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/merkle"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	latestVersionKey      = "s/latest"
	commitInfoKeyFmt      = "s/%d" // s/<version>
	defaultRootNumHistory = 2
)

// rootMultiStore is composed of many CommitStores.
// Name contrasts with cacheMultiStore which is for cache-wrapping
// other MultiStores.
// Implements MultiStore.
type rootMultiStore struct {
	db            dbm.DB
	lastCommitID  CommitID
	storesParams  map[StoreKey]storeParams
	stores        map[StoreKey]CommitStore
	keysByName    map[string]StoreKey
	opsNumHistory int
	opsMaps       map[int64]map[string][]merkle.Op
}

var _ CommitMultiStore = (*rootMultiStore)(nil)
var _ Queryable = (*rootMultiStore)(nil)

// nolint
func NewCommitMultiStore(db dbm.DB) *rootMultiStore {
	return &rootMultiStore{
		db:            db,
		storesParams:  make(map[StoreKey]storeParams),
		stores:        make(map[StoreKey]CommitStore),
		keysByName:    make(map[string]StoreKey),
		opsNumHistory: defaultRootNumHistory,
		opsMaps:       make(map[int64]map[string][]merkle.Op),
	}
}

// Implements Store.
func (rs *rootMultiStore) GetStoreType() StoreType {
	return sdk.StoreTypeMulti
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) MountStoreWithDB(key StoreKey, typ StoreType, db dbm.DB) {
	if key == nil {
		panic("MountIAVLStore() key cannot be nil")
	}
	if _, ok := rs.storesParams[key]; ok {
		panic(fmt.Sprintf("rootMultiStore duplicate store key %v", key))
	}
	rs.storesParams[key] = storeParams{
		key: key,
		typ: typ,
		db:  db,
	}
	rs.keysByName[key.Name()] = key
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) GetCommitStore(key StoreKey) CommitStore {
	return rs.stores[key]
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) GetCommitKVStore(key StoreKey) CommitKVStore {
	return rs.stores[key].(CommitKVStore)
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) LoadLatestVersion() error {
	ver := getLatestVersion(rs.db)
	return rs.LoadVersion(ver)
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) LoadVersion(ver int64) error {

	// Special logic for version 0
	if ver == 0 {
		for key, storeParams := range rs.storesParams {
			id := CommitID{}
			store, err := rs.loadCommitStoreFromParams(id, storeParams)
			if err != nil {
				return fmt.Errorf("Failed to load rootMultiStore: %v", err)
			}
			rs.stores[key] = store
		}

		rs.lastCommitID = CommitID{}
		return nil
	}
	// Otherwise, version is 1 or greater

	// Get commitInfo
	cInfo, err := getCommitInfo(rs.db, ver)
	if err != nil {
		return err
	}

	// Load each Store
	var newStores = make(map[StoreKey]CommitStore)
	for _, storeInfo := range cInfo.StoreInfos {
		key, commitID := rs.nameToKey(storeInfo.Name), storeInfo.Core.CommitID
		storeParams := rs.storesParams[key]
		store, err := rs.loadCommitStoreFromParams(commitID, storeParams)
		if err != nil {
			return fmt.Errorf("Failed to load rootMultiStore: %v", err)
		}
		newStores[key] = store
	}

	// If any CommitStoreLoaders were not used, return error.
	for key := range rs.storesParams {
		if _, ok := newStores[key]; !ok {
			return fmt.Errorf("Unused CommitStoreLoader: %v", key)
		}
	}

	// Success.
	rs.lastCommitID = cInfo.CommitID()
	rs.stores = newStores
	return nil
}

//----------------------------------------
// +CommitStore

// Implements Committer/CommitStore.
func (rs *rootMultiStore) LastCommitID() CommitID {
	return rs.lastCommitID
}

// Implements Committer/CommitStore.
func (rs *rootMultiStore) Commit() CommitID {

	// Commit stores.
	version := rs.lastCommitID.Version + 1
	commitInfo := commitStores(version, rs.stores)

	// Need to update atomically.
	batch := rs.db.NewBatch()
	setCommitInfo(batch, version, commitInfo)
	setLatestVersion(batch, version)
	batch.Write()

	// Cache subproof ops
	rs.opsMaps[version] = commitInfo.Proofs()
	if len(rs.opsMaps) > rs.opsNumHistory {
		delete(rs.opsMaps, version-int64(rs.opsNumHistory))
	}

	// Prepare for next version.
	commitID := CommitID{
		Version: version,
		Hash:    commitInfo.Hash(),
	}
	rs.lastCommitID = commitID
	return commitID
}

// Implements CacheWrapper/Store/CommitStore.
func (rs *rootMultiStore) CacheWrap() CacheWrap {
	return rs.CacheMultiStore().(CacheWrap)
}

//----------------------------------------
// +MultiStore

// Implements MultiStore.
func (rs *rootMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStoreFromRMS(rs)
}

// Implements MultiStore.
func (rs *rootMultiStore) GetStore(key StoreKey) Store {
	return rs.stores[key]
}

// Implements MultiStore.
func (rs *rootMultiStore) GetKVStore(key StoreKey) KVStore {
	return rs.stores[key].(KVStore)
}

// Implements MultiStore.
func (rs *rootMultiStore) GetKVStoreWithGas(meter sdk.GasMeter, key StoreKey) KVStore {
	return NewGasKVStore(meter, rs.GetKVStore(key))
}

// getStoreByName will first convert the original name to
// a special key, before looking up the CommitStore.
// This is not exposed to the extensions (which will need the
// StoreKey), but is useful in main, and particularly app.Query,
// in order to convert human strings into CommitStores.
func (rs *rootMultiStore) getStoreByName(name string) Store {
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
func (rs *rootMultiStore) Query(req abci.RequestQuery) (abci.ResponseQuery, merkle.Proof) {
	// Query just routes this to a substore.
	path := req.Path
	storeName, subpath, err := parsePath(path)
	if err != nil {
		return err.QueryResult(), nil
	}

	store := rs.getStoreByName(storeName)
	if store == nil {
		msg := fmt.Sprintf("no such store: %s", storeName)
		return sdk.ErrUnknownRequest(msg).QueryResult(), nil
	}
	queryable, ok := store.(Queryable)
	if !ok {
		msg := fmt.Sprintf("store %s doesn't support queries", storeName)
		return sdk.ErrUnknownRequest(msg).QueryResult(), nil
	}

	// trim the path and make the query
	req.Path = subpath
	res, prf := queryable.Query(req)

	if req.Prove && prf != nil {
		opsMap, ok := rs.opsMaps[res.Height]
		if !ok {
			ci, err := getCommitInfo(rs.db, res.Height)
			if err != nil {
				return sdk.ErrInternal(err.Error()).QueryResult(), nil
			}
			opsMap = ci.Proofs()
		}
		prf = append(prf, opsMap[storeName]...)
	}

	return res, prf
}

// parsePath expects a format like /<storeName>[/<subpath>]
// Must start with /, subpath may be empty
// Returns error if it doesn't start with /
func parsePath(path string) (storeName string, subpath string, err sdk.Error) {
	if !strings.HasPrefix(path, "/") {
		err = sdk.ErrUnknownRequest(fmt.Sprintf("invalid path: %s", path))
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

func (rs *rootMultiStore) loadCommitStoreFromParams(id CommitID, params storeParams) (store CommitStore, err error) {
	var db dbm.DB
	if params.db != nil {
		db = dbm.NewPrefixDB(params.db, []byte("s/_/"))
	} else {
		db = dbm.NewPrefixDB(rs.db, []byte("s/k:"+params.key.Name()+"/"))
	}
	switch params.typ {
	case sdk.StoreTypeMulti:
		panic("recursive MultiStores not yet supported")
		// TODO: id?
		// return NewCommitMultiStore(db, id)
	case sdk.StoreTypeIAVL:
		store, err = LoadIAVLStore(db, id)
		return
	case sdk.StoreTypeDB:
		panic("dbm.DB is not a CommitStore")
	default:
		panic(fmt.Sprintf("unrecognized store type %v", params.typ))
	}
}

func (rs *rootMultiStore) nameToKey(name string) StoreKey {
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
	key StoreKey
	db  dbm.DB
	typ StoreType
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
	// TODO cache to ci.hash []byte
	m := make(map[string]libmerkle.Hasher, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo
	}
	return libmerkle.SimpleHashFromMap(m)
}

func (ci commitInfo) CommitID() CommitID {
	return CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

func (ci commitInfo) Proofs() (res map[string][]merkle.Op) {
	m := make(map[string]libmerkle.Hasher, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo
	}
	_, proofs, keys := libmerkle.SimpleProofsFromMap(m)

	res = make(map[string][]merkle.Op)
	for i, key := range keys {
		si := m[key].(storeInfo)
		res[key] = []merkle.Op{
			RootMultistoreOp{si.Name, si.Core.CommitID.Version},
			merkle.FromSimpleProof(proofs[key], i, len(keys)),
		}
	}
	return
}

//----------------------------------------
// storeInfo

// storeInfo contains the name and core reference for an
// underlying store.  It is the leaf of the rootMultiStores top
// level simple merkle tree.
type storeInfo struct {
	Name string
	Core storeCore
}

type storeCore struct {
	// StoreType StoreType
	CommitID CommitID
	// ... maybe add more state
}

// Implements merkle.Hasher.
func (si storeInfo) Hash() []byte {
	// Doesn't write Name, since merkle.SimpleHashFromMap() will
	// include them via the keys.
	bz, _ := cdc.MarshalBinary(si.Core) // Does not error
	hasher := ripemd160.New()
	hasher.Write(bz)
	return hasher.Sum(nil)
}

type RootMultistoreOp struct {
	Name    string
	Version int64
}

const RootMultistoreOpType = merkle.OpType("rootmultistore")

func (op RootMultistoreOp) Run(value [][]byte) ([][]byte, error) {
	if len(value) != 1 {
		return nil, fmt.Errorf("Value size is not 1")
	}

	si := storeInfo{
		Name: op.Name,
		Core: storeCore{
			CommitID: CommitID{
				Version: op.Version,
				Hash:    value[0],
			},
		},
	}
	kvp := libmerkle.KVPair{[]byte(op.Name), si.Hash()}
	return [][]byte{kvp.Hash()}, nil
}

func (op RootMultistoreOp) GetKey() string {
	return op.Name
}

func (op RootMultistoreOp) Raw() (res merkle.RawOp, err error) {
	bz, err := json.Marshal(op)
	if err != nil {
		return
	}

	return merkle.RawOp{
		Type: RootMultistoreOpType,
		Data: string(bz),
		Key:  op.Name,
	}, nil
}

func WrapOpDecoder(decode merkle.OpDecoder) merkle.OpDecoder {
	return func(ro merkle.RawOp) (res merkle.Op, err error) {
		switch ro.Type {
		case RootMultistoreOpType:
			var op RootMultistoreOp
			err = json.Unmarshal([]byte(ro.Data), &op)
			res = op
		default:
			return decode(ro)
		}
		return
	}
}

//----------------------------------------
// Misc.

func getLatestVersion(db dbm.DB) int64 {
	var latest int64
	latestBytes := db.Get([]byte(latestVersionKey))
	if latestBytes == nil {
		return 0
	}
	err := cdc.UnmarshalBinary(latestBytes, &latest)
	if err != nil {
		panic(err)
	}
	return latest
}

// Set the latest version.
func setLatestVersion(batch dbm.Batch, version int64) {
	latestBytes, _ := cdc.MarshalBinary(version) // Does not error
	batch.Set([]byte(latestVersionKey), latestBytes)
}

// Commits each store and returns a new commitInfo.
func commitStores(version int64, storeMap map[StoreKey]CommitStore) commitInfo {
	storeInfos := make([]storeInfo, 0, len(storeMap))

	for key, store := range storeMap {
		// Commit
		commitID := store.Commit()

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
		return commitInfo{}, fmt.Errorf("Failed to get rootMultiStore: no data")
	}

	// Parse bytes.
	var cInfo commitInfo
	err := cdc.UnmarshalBinary(cInfoBytes, &cInfo)
	if err != nil {
		return commitInfo{}, fmt.Errorf("Failed to get rootMultiStore: %v", err)
	}
	return cInfo, nil
}

// Set a commitInfo for given version.
func setCommitInfo(batch dbm.Batch, version int64, cInfo commitInfo) {
	cInfoBytes, err := cdc.MarshalBinary(cInfo)
	if err != nil {
		panic(err)
	}
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, version)
	batch.Set([]byte(cInfoKey), cInfoBytes)
}
