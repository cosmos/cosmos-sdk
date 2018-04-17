package store

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ripemd160"

	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	libmerkle "github.com/tendermint/tmlibs/merkle"

	"github.com/cosmos/cosmos-sdk/merkle"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>
)

// rootMultiStore is composed of many CommitStores.
// Name contrasts with cacheMultiStore which is for cache-wrapping
// other MultiStores.
// Implements MultiStore.
type rootMultiStore struct {
	db           dbm.DB
	lastCommitID CommitID
	storesParams map[StoreKey]storeParams
	stores       map[StoreKey]CommitStore
	keysByName   map[string]StoreKey
	indexByName  map[string]int

	// cached proofs
	proofs []merkle.ExistsProof
}

var _ CommitMultiStore = (*rootMultiStore)(nil)
var _ Queryable = (*rootMultiStore)(nil)

func NewCommitMultiStore(db dbm.DB) *rootMultiStore {
	return &rootMultiStore{
		db:           db,
		storesParams: make(map[StoreKey]storeParams),
		stores:       make(map[StoreKey]CommitStore),
		keysByName:   make(map[string]StoreKey),
		indexByName:  make(map[string]int),
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
		db:  db,
		typ: typ,
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

	// Prepare for next version.
	commitID := CommitID{
		Version: version,
		Hash:    commitInfo.Hash(),
	}
	rs.lastCommitID = commitID

	// cache substore proofs
	rs.cacheSubstoreProofs(commitInfo.StoreInfos)
	rs.setIndexByName(commitInfo.StoreInfos)

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
func (rs *rootMultiStore) Query(req abci.RequestQuery) (value []byte, proof merkle.MultiProof, err error) {
	// Query just routes this to a substore.
	path := req.Path
	storeName, subpath, err := parsePath(path)
	if err != nil {
		return
	}

	store := rs.getStoreByName(storeName)
	if store == nil {
		msg := fmt.Sprintf("no such store: %s", storeName)
		err = sdk.ErrUnknownRequest(msg)
		return
	}
	queryable, ok := store.(Queryable)
	if !ok {
		msg := fmt.Sprintf("store %s doesn't support queries", storeName)
		err = sdk.ErrUnknownRequest(msg)
		return
	}

	// trim the path and make the query
	req.Path = subpath
	value, proof, err = queryable.Query(req)
	if err != nil {
		return
	}

	subproof := rs.proofs[rs.indexByName[storeName]]
	proof.SubProofs = append(proof.SubProofs, subproof)

	return
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
	db := rs.db
	if params.db != nil {
		db = params.db
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
	for key, _ := range rs.storesParams {
		if key.Name() == name {
			return key
		}
	}
	panic("Unknown name " + name)
}

func makeInfoMap(infos []storeInfo) map[string]libmerkle.Hasher {
	m := make(map[string]libmerkle.Hasher, len(infos))
	for _, info := range infos {
		m[info.Name] = info
	}
	return m
}

func (rs *rootMultiStore) setIndexByName(infos []storeInfo) {
	rs.indexByName = make(map[string]int)
	m := makeInfoMap(infos)
	sm := libmerkle.NewSimpleMap()
	for k, v := range m {
		sm.Set(k, v)
	}

	kvs := sm.KVPairs()

	var key [20]byte
	kvm := make(map[[20]byte]int)

	for i, kvp := range kvs {
		copy(key[:], kvp.Key[:20]) // Size of a RIPEMD160 hash is 20
		kvm[key] = i
	}

	for _, info := range infos {
		hash := libmerkle.SimpleHashFromBytes([]byte(info.Name))
		copy(key[:], hash[:20])
		rs.indexByName[info.Name] = kvm[key]
	}
}

func (rs *rootMultiStore) cacheSubstoreProofs(infos []storeInfo) {
	m := makeInfoMap(infos)
	_, proofs := libmerkle.SimpleProofsFromMap(m)

	rs.proofs = make([]merkle.ExistsProof, len(proofs))
	for i, p := range proofs {
		proof, err := merkle.FromSimpleProof(p, i, len(proofs))
		if err != nil {
			panic(err)
		}
		rs.proofs[i] = proof
	}
}

//----------------------------------------
// storeParams

type storeParams struct {
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

	return commitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
}

// same with commitStores but use LastCommitID instead of Commit
func getCommitIDStores(storeMap map[StoreKey]CommitStore) []storeInfo {
	storeInfos := make([]storeInfo, 0, len(storeMap))

	for key, store := range storeMap {
		commitID := store.LastCommitID()

		si := storeInfo{}
		si.Name = key.Name()
		si.Core.CommitID = commitID

		storeInfos = append(storeInfos, si)
	}

	return storeInfos
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
