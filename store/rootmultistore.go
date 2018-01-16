package store

import (
	"fmt"

	"github.com/tendermint/go-wire"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
	"golang.org/x/crypto/ripemd160"
)

const (
	latestVersionKey  = "s/latest"
	commitStateKeyFmt = "s/%d" // s/<version>
)

// rootMultiStore is composed of many CommitStores.
// Name contrasts with cacheMultiStore which is for cache-wrapping
// other MultiStores.
// Implements MultiStore.
type rootMultiStore struct {
	db           dbm.DB
	nextVersion  int64
	lastCommitID CommitID
	storeLoaders map[SubstoreKey]CommitStoreLoader
	substores    map[SubstoreKey]CommitStore
}

var _ CommitMultiStore = (*rootMultiStore)(nil)

func NewCommitMultiStore(db dbm.DB) *rootMultiStore {
	return &rootMultiStore{
		db:           db,
		nextVersion:  0,
		storeLoaders: make(map[SubstoreKey]CommitStoreLoader),
		substores:    make(map[SubstoreKey]CommitStore),
	}
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) SetSubstoreLoader(key SubstoreKey, loader CommitStoreLoader) {
	if _, ok := rs.storeLoaders[key]; ok {
		panic(fmt.Sprintf("rootMultiStore duplicate substore key", key))
	}
	rs.storeLoaders[key] = loader
}

// Implements CommitMultiStore.
func (rs *rootMultiStore) GetSubstore(key SubstoreKey) CommitStore {
	return rs.substores[key]
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
		for key, storeLoader := range rs.storeLoaders {
			store, err := storeLoader(CommitID{})
			if err != nil {
				return fmt.Errorf("Failed to load rootMultiStore: %v", err)
			}
			rs.substores[key] = store
		}

		rs.nextVersion = 1
		rs.lastCommitID = CommitID{}
		return nil
	}
	// Otherwise, version is 1 or greater

	// Get commitState
	state, err := getCommitState(rs.db, ver)
	if err != nil {
		return err
	}

	// Load each Substore
	var newSubstores = make(map[SubstoreKey]CommitStore)
	for _, store := range state.Substores {
		key, commitID := rs.nameToKey(store.Name), store.CommitID
		storeLoader := rs.storeLoaders[key]
		if storeLoader == nil {
			return fmt.Errorf("Failed to load rootMultiStore substore %v for commitID %v: %v", key, commitID, err)
		}
		store, err := storeLoader(commitID)
		if err != nil {
			return fmt.Errorf("Failed to load rootMultiStore: %v", err)
		}
		newSubstores[key] = store
	}

	// If any CommitStoreLoaders were not used, return error.
	for key := range rs.storeLoaders {
		if _, ok := newSubstores[key]; !ok {
			return fmt.Errorf("Unused CommitStoreLoader: %v", key)
		}
	}

	// Success.
	rs.nextVersion = ver + 1
	rs.lastCommitID = state.CommitID()
	rs.substores = newSubstores
	return nil
}

func (rs *rootMultiStore) nameToKey(name string) SubstoreKey {
	for key, _ := range rs.storeLoaders {
		if key.Name() == name {
			return key
		}
	}
	panic("Unknown name " + name)
}

//----------------------------------------
// +CommitStore

// Implements CommitStore.
func (rs *rootMultiStore) Commit() CommitID {

	// Commit substores.
	version := rs.nextVersion
	state := commitSubstores(version, rs.substores)

	// Need to update self state atomically.
	batch := rs.db.NewBatch()
	setCommitState(batch, version, state)
	setLatestVersion(batch, version)
	batch.Write()

	// Prepare for next version.
	rs.nextVersion = version + 1
	commitID := CommitID{
		Version: version,
		Hash:    state.Hash(),
	}
	rs.lastCommitID = commitID
	return commitID
}

// Implements CommitStore.
func (rs *rootMultiStore) CacheWrap() CacheWrap {
	return rs.CacheMultiStore().(CacheWrap)
}

//----------------------------------------
// +MultiStore

// Implements MultiStore.
func (rs *rootMultiStore) LastCommitID() CommitID {
	return rs.lastCommitID
}

// Implements MultiStore.
// NOTE: Returns 0 unless LoadVersion() or LoadLatestVersion() is called.
func (rs *rootMultiStore) NextVersion() int64 {
	return rs.nextVersion
}

// Implements MultiStore.
func (rs *rootMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStoreFromRMS(rs)
}

// Implements MultiStore.
func (rs *rootMultiStore) GetStore(key SubstoreKey) interface{} {
	return rs.substores[key]
}

// Implements MultiStore.
func (rs *rootMultiStore) GetKVStore(key SubstoreKey) KVStore {
	return rs.substores[key].(KVStore)
}

//----------------------------------------
// commitState

// NOTE: Keep commitState a simple immutable struct.
type commitState struct {

	// Version
	Version int64

	// Substore info for
	Substores []substore
}

// Hash returns the simple merkle root hash of the substores sorted by name.
func (cs commitState) Hash() []byte {
	// TODO cache to cs.hash []byte
	m := make(map[string]interface{}, len(cs.Substores))
	for _, substore := range cs.Substores {
		m[substore.Name] = substore
	}
	return merkle.SimpleHashFromMap(m)
}

func (cs commitState) CommitID() CommitID {
	return CommitID{
		Version: cs.Version,
		Hash:    cs.Hash(),
	}
}

//----------------------------------------
// substore state

// substore contains the name and core reference for an underlying store.
// It is the leaf of the rootMultiStores top level simple merkle tree.
type substore struct {
	Name string
	substoreCore
}

type substoreCore struct {
	CommitID CommitID
	// ... maybe add more state
}

// Hash returns the RIPEMD160 of the wire-encoded substore.
func (sc substoreCore) Hash() []byte {
	scBytes, _ := wire.MarshalBinary(sc) // Does not error
	hasher := ripemd160.New()
	hasher.Write(scBytes)
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
	err := wire.UnmarshalBinary(latestBytes, &latest)
	if err != nil {
		panic(err)
	}
	return latest
}

// Set the latest version.
func setLatestVersion(batch dbm.Batch, version int64) {
	latestBytes, _ := wire.MarshalBinary(version) // Does not error
	batch.Set([]byte(latestVersionKey), latestBytes)
}

// Commits each substore and returns a new commitState.
func commitSubstores(version int64, substoresMap map[SubstoreKey]CommitStore) commitState {
	substores := make([]substore, 0, len(substoresMap))

	for key, store := range substoresMap {
		// Commit
		commitID := store.Commit()

		// Record CommitID
		substore := substore{}
		substore.Name = key.Name()
		substore.CommitID = commitID
		substores = append(substores, substore)
	}

	return commitState{
		Version:   version,
		Substores: substores,
	}
}

// Gets commitState from disk.
func getCommitState(db dbm.DB, ver int64) (commitState, error) {

	// Get from DB.
	commitStateKey := fmt.Sprintf(commitStateKeyFmt, ver)
	stateBytes := db.Get([]byte(commitStateKey))
	if stateBytes == nil {
		return commitState{}, fmt.Errorf("Failed to get rootMultiStore: no data")
	}

	// Parse bytes.
	var state commitState
	err := wire.UnmarshalBinary(stateBytes, &state)
	if err != nil {
		return commitState{}, fmt.Errorf("Failed to get rootMultiStore: %v", err)
	}
	return state, nil
}

// Set a commit state for given version.
func setCommitState(batch dbm.Batch, version int64, state commitState) {
	stateBytes, err := wire.MarshalBinary(state)
	if err != nil {
		panic(err)
	}
	commitStateKey := fmt.Sprintf(commitStateKeyFmt, version)
	batch.Set([]byte(commitStateKey), stateBytes)
}
