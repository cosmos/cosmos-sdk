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
	curVersion   int64
	lastCommitID CommitID
	storeLoaders map[string]CommitStoreLoader
	substores    map[string]CommitStore
}

func NewMultiStore(db dbm.DB) *rootMultiStore {
	return &rootMultiStore{
		db:           db,
		curVersion:   0,
		storeLoaders: make(map[string]CommitStoreLoader),
		substores:    make(map[string]CommitStore),
	}
}

func (rs *rootMultiStore) SetSubstoreLoader(name string, loader CommitStoreLoader) {
	if _, ok := rs.storeLoaders[name]; ok {
		panic(fmt.Sprintf("rootMultiStore duplicate substore name " + name))
	}
	rs.storeLoaders[name] = loader
}

// Call once after all calls to SetSubstoreLoader are complete.
func (rs *rootMultiStore) LoadLatestVersion() error {
	ver := getLatestVersion(rs.db)
	return rs.LoadVersion(ver)
}

// NOTE: Returns 0 unless LoadVersion() or LoadLatestVersion() is called.
func (rs *rootMultiStore) GetCurrentVersion() int64 {
	return rs.curVersion
}

func (rs *rootMultiStore) LoadVersion(ver int64) error {

	// Special logic for version 0
	if ver == 0 {
		for name, storeLoader := range rs.storeLoaders {
			store, err := storeLoader(CommitID{})
			if err != nil {
				return fmt.Errorf("Failed to load rootMultiStore: %v", err)
			}
			rs.substores[name] = store
		}

		rs.curVersion = 1
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
	var newSubstores = make(map[string]CommitStore)
	for _, store := range state.Substores {
		name, commitID := store.Name, store.CommitID
		storeLoader := rs.storeLoaders[name]
		if storeLoader == nil {
			return fmt.Errorf("Failed to load rootMultiStore substore %v for commitID %v: %v", name, commitID, err)
		}
		store, err := storeLoader(commitID)
		if err != nil {
			return fmt.Errorf("Failed to load rootMultiStore: %v", err)
		}
		newSubstores[name] = store
	}

	// If any CommitStoreLoaders were not used, return error.
	for name := range rs.storeLoaders {
		if _, ok := newSubstores[name]; !ok {
			return fmt.Errorf("Unused CommitStoreLoader: %v", name)
		}
	}

	// Success.
	rs.curVersion = ver + 1
	rs.lastCommitID = state.CommitID()
	rs.substores = newSubstores
	return nil
}

// Implements CommitStore
func (rs *rootMultiStore) Commit() CommitID {

	// Commit substores
	version := rs.curVersion
	state := commitSubstores(version, rs.substores)

	// Need to update self state atomically.
	batch := rs.db.NewBatch()
	setCommitState(batch, version, state)
	setLatestVersion(batch, version)
	batch.Write()

	// Prepare for next version.
	rs.curVersion = version + 1
	commitID := CommitID{
		Version: version,
		Hash:    state.Hash(),
	}
	rs.lastCommitID = commitID
	return commitID
}

// Implements CommitStore
func (rs *rootMultiStore) CacheWrap() CacheWrap {
	return rs.CacheMultiStore().(CacheWrap)
}

// Get the last committed CommitID
func (rs *rootMultiStore) LastCommitID() CommitID {
	return rs.lastCommitID
}

// Implements MultiStore
func (rs *rootMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStoreFromRMS(rs)
}

// Implements MultiStore
func (rs *rootMultiStore) GetCommitStore(name string) CommitStore {
	return rs.substores[name]
}

// Implements MultiStore
func (rs *rootMultiStore) GetKVStore(name string) KVStore {
	return rs.substores[name].(KVStore)
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
	scBytes, _ := wire.Marshal(sc) // Does not error
	hasher := ripemd160.New()
	hasher.Write(scBytes)
	return hasher.Sum(nil)
}

//----------------------------------------

func getLatestVersion(db dbm.DB) int64 {
	var latest int64
	latestBytes := db.Get([]byte(latestVersionKey))
	if latestBytes == nil {
		return 0
	}
	err := wire.Unmarshal(latestBytes, &latest)
	if err != nil {
		panic(err)
	}
	return latest
}

// Set the latest version.
func setLatestVersion(batch dbm.Batch, version int64) {
	latestBytes, _ := wire.Marshal(version) // Does not error
	batch.Set([]byte(latestVersionKey), latestBytes)
}

// Commits each substore and returns a new commitState.
func commitSubstores(version int64, substoresMap map[string]CommitStore) commitState {
	substores := make([]substore, 0, len(substoresMap))

	for name, store := range substoresMap {
		// Commit
		commitID := store.Commit()

		// Record CommitID
		substore := substore{}
		substore.Name = name
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
	err := wire.Unmarshal(stateBytes, &state)
	if err != nil {
		return commitState{}, fmt.Errorf("Failed to get rootMultiStore: %v", err)
	}
	return state, nil
}

// Set a commit state for given version.
func setCommitState(batch dbm.Batch, version int64, state commitState) {
	stateBytes, err := wire.Marshal(state)
	if err != nil {
		panic(err)
	}
	commitStateKey := fmt.Sprintf(commitStateKeyFmt, version)
	batch.Set([]byte(commitStateKey), stateBytes)
}
