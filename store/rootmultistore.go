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

func (rs *rootMultiStore) SetCommitStoreLoader(name string, loader CommitStoreLoader) {
	if _, ok := rs.storeLoaders[name]; ok {
		panic(fmt.Sprintf("rootMultiStore duplicate substore name " + name))
	}
	rs.storeLoaders[name] = loader
}

// Call once after all calls to SetCommitStoreLoader are complete.
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
			rs.curVersion = 1
			rs.lastCommitID = CommitID{}
			rs.substores[name] = store
		}
		return nil
	}
	// Otherwise, version is 1 or greater

	// Load commitState
	state, err := loadCommitState(rs.db, ver)
	if err != nil {
		return err
	}

	// Load each Substore
	var newSubstores = make(map[string]CommitStore)
	for _, store := range state.Substores {
		name, commitID := store.Name, store.CommitID
		storeLoader := rs.storeLoaders[name]
		if storeLoader == nil {
			return fmt.Errorf("Failed to loadrootMultiStore: CommitStoreLoader missing for %v", name)
		}
		store, err := storeLoader(commitID)
		if err != nil {
			return fmt.Errorf("Failed to load rootMultiStore: %v", err)
		}
		newSubstores[name] = store
	}

	// If any CommitStoreLoaders were not used, return error.
	for name := range rs.storeLoaders {
		if _, ok := rs.substores[name]; !ok {
			return fmt.Errorf("Unused CommitStoreLoader: %v", name)
		}
	}

	// Success.
	rs.curVersion = ver + 1
	rs.lastCommitID = state.CommitID()
	rs.substores = newSubstores
	return nil
}

// Commits each substore and gets commitState.
func (rs *rootMultiStore) doCommit() commitState {
	version := rs.curVersion
	substores := make([]substore, len(rs.substores))

	for name, store := range rs.substores {
		// Commit
		commitID := store.Commit()

		// Record CommitID
		substore := substore{}
		substore.Name = name
		substore.CommitID = commitID
		substores = append(substores, substore)
	}

	// Incr curVersion
	rs.curVersion += 1

	return commitState{
		Version:   version,
		Substores: substores,
	}
}

//----------------------------------------

// Implements CommitStore
func (rs *rootMultiStore) Commit() CommitID {

	version := rs.curVersion

	// Needs to be transactional
	batch := rs.db.NewBatch()

	// Commit each substore and get commitState
	state := rs.doCommit()
	stateBytes, err := wire.Marshal(state)
	if err != nil {
		panic(err)
	}
	commitStateKey := fmt.Sprintf(commitStateKeyFmt, rs.curVersion)
	batch.Set([]byte(commitStateKey), stateBytes)

	// Save the latest version
	latestBytes, _ := wire.Marshal(rs.curVersion) // Does not error
	batch.Set([]byte(latestVersionKey), latestBytes)

	batch.Write()
	rs.curVersion += 1
	commitID := CommitID{
		Version: version,
		Hash:    state.Hash(),
	}
	rs.lastCommitID = commitID
	return commitID
}

// Implements CommitStore
func (rs *rootMultiStore) CacheWrap() CacheWrap {
	return rs.CacheMultiStore()
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

// Implements MultiStore
func (rs *rootMultiStore) GetIterKVStore(name string) IterKVStore {
	return rs.substores[name].(IterKVStore)
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

// loads commitState from disk.
func loadCommitState(db dbm.DB, ver int64) (commitState, error) {

	// Load from DB.
	commitStateKey := fmt.Sprintf(commitStateKeyFmt, ver)
	stateBytes := db.Get([]byte(commitStateKey))
	if stateBytes == nil {
		return commitState{}, fmt.Errorf("Failed to load rootMultiStore: no data")
	}

	// Parse bytes.
	var state commitState
	err := wire.Unmarshal(stateBytes, &state)
	if err != nil {
		return commitState{}, fmt.Errorf("Failed to load rootMultiStore: %v", err)
	}
	return state, nil
}

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

type substore struct {
	Name string
	substoreCore
}

type substoreCore struct {
	CommitID CommitID
	// ... maybe add more state
}

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
