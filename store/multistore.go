package store

import (
	"fmt"
	"sort"

	"github.com/tendermint/go-wire"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/merkle"
	"golang.org/x/crypto/ripemd160"
)

const (
	msLatestKey   = "s/latest"
	msStateKeyFmt = "s/%d" // s/<version>
)

type MultiStore interface {
	CacheWrappable

	// Convenience
	GetStore(name string) interface{}
	GetKVStore(name string) KVStore
	GetIterKVStore(name string) IterKVStore
}

//----------------------------------------

// RootStore is composed of many Committers.
// Implements MultiStore.
type RootStore struct {
	db           dbm.DB
	version      uint64
	storeLoaders map[string]CommitterLoader
	substores    map[string]Committer
}

func NewRootStore(db dbm.DB) *RootStore {
	return &RootStore{
		db:           db,
		version:      0,
		storeLoaders: make(map[string]CommitterLoader),
		substores:    make(map[string]Committer),
	}
}

func (rs *RootStore) SetCommitterLoader(name string, loader CommitterLoader) {
	if _, ok := rs.storeLoaders[name]; ok {
		panic(fmt.Sprintf("RootStore duplicate substore name " + name))
	}
	rs.storeLoaders[name] = loader
}

//----------------------------------------
// RootStore state

type msState struct {
	Substores []substore
}

func (rs *msState) Sort() {
	rs.Substores.Sort()
}

func (rs *msState) Hash() []byte {
	m := make(map[string]interface{}, len(rs.Substores))
	for _, substore := range rs.Substores {
		m[substore.name] = substore.ssState
	}
	return merkle.SimpleHashFromMap(m)
}

//----------------------------------------
// substore state

type substore struct {
	name string
	ssState
}

// This gets serialized by go-wire
type ssState struct {
	CommitID CommitID
	// ... maybe add more state
}

func (ss ssState) Hash() []byte {
	ssBytes, _ := wire.Marshal(ss) // Does not error
	hasher := ripemd160.New()
	hasher.Write(ssBytes)
	return hasher.Sum(nil)
}

//----------------------------------------

// Call once after all calls to SetStoreLoader are complete.
func (rs *RootStore) LoadLatestVersion() error {
	ver := rs.getLatestVersion()
	rs.LoadVersion(ver)
}

func (rs *RootStore) getLatestVersion() uint64 {
	var latest uint64
	latestBytes := rs.db.Get(msLatestKey)
	if latestBytes == nil {
		return 0
	}
	err := wire.Unmarshal(latestBytes, &latest)
	if err != nil {
		panic(err)
	}
	return latest
}

func (rs *RootStore) LoadVersion(ver uint64) error {
	rs.version = ver

	// Special logic for version 0
	if ver == 0 {
		for name, storeLoader := range rs.storeLoaders {
			store, err := storeLoader(CommitID{Version: 0})
			if err != nil {
				return fmt.Errorf("Failed to load RootStore: %v", err)
			}
			rs.substores[name] = store
		}
		return nil
	}
	// Otherwise, version is 1 or greater

	msStateKey := fmt.Sprintf(msStateKeyFmt, ver)
	stateBytes := rs.db.Get(msStateKey, ver)
	if bz == nil {
		return fmt.Errorf("Failed to load RootStore: no data")
	}
	var state msState
	err := wire.Unmarshal(stateBytes, &state)
	if err != nil {
		return fmt.Errorf("Failed to load RootStore: %v", err)
	}

	// Load each Substore
	for _, store := range state.Substores {
		name, commitID := store.Name, store.CommitID
		storeLoader := rs.storeLoaders[name]
		if storeLoader == nil {
			return fmt.Errorf("Failed to loadRootStore: StoreLoader missing for %v", name)
		}
		store, err := storeLoader(commitID)
		if err != nil {
			return fmt.Errorf("Failed to load RootStore: %v", err)
		}
		rs.substores[name] = store
	}

	// If any StoreLoaders were not used, return error.
	for name := range rs.storeLoaders {
		if _, ok := rs.substores[name]; !ok {
			return fmt.Errorf("Unused StoreLoader: %v", name)
		}
	}

	return nil
}

// Implements Committer
func (rs *RootStore) Commit() CommitID {

	// Needs to be transactional
	batch := rs.db.NewBatch()

	// Save msState
	var state msState
	for name, store := range rs.substores {
		commitID := store.Commit()
		state.Substores = append(state.Substores,
			ssState{
				Name:     name,
				CommitID: commitID,
			},
		)
	}
	state.Sort()
	stateBytes, err := wire.Marshal(state)
	if err != nil {
		panic(err)
	}
	msStateKey := fmt.Sprintf(msStateKeyFmt, rs.version)
	batch.Set(msStateKey, stateBytes)

	// Save msLatest
	latestBytes, _ := wire.Marshal(rs.version) // Does not error
	batch.Set(msLatestKey, latestBytes)

	batch.Write()
	batch.version += 1
}

// Implements MultiStore/CacheWrappable
func (rs *RootStore) CacheWrap() (o interface{}) {
	return newCacheMultiStore(rs)
}

// Implements MultiStore/CacheWrappable
func (rs *RootStore) GetCommitter(name string) Committer {
	return rs.store[name]
}

// Implements MultiStore/CacheWrappable
func (rs *RootStore) GetKVStore(name string) KVStore {
	return rs.store[name].(KVStore)
}

// Implements MultiStore/CacheWrappable
func (rs *RootStore) GetIterKVStore(name string) IterKVStore {
	return rs.store[name].(IterKVStore)
}

//----------------------------------------
// ssStates

type ssStates []ssState

func (ssz ssStates) Len() int           { return len(ssz) }
func (ssz ssStates) Less(i, j int) bool { return ssz[i].Key < ssz[j].Key }
func (ssz ssStates) Swap(i, j int)      { ssz[i], ssz[j] = ssz[j], ssz[i] }
func (ssz ssStates) Sort()              { sort.Sort(ssz) }

func (ssz ssStates) Hash() []byte {
	hz := make([]merkle.Hashable, len(ssz))
	for i, ss := range ssz {
		hz[i] = ss
	}
	return merkle.SimpleHashFromHashables(hz)
}

//----------------------------------------
// cacheMultiStore

type cwWriter interface {
	Write()
}

// cacheMultiStore holds many CacheWrap'd stores.
// Implements MultiStore.
type cacheMultiStore struct {
	db        dbm.DB
	version   uint64
	substores map[string]cwWriter
}

func newCacheMultiStore(rs *RootStore) MultiStore {
	cms := cacheMultiStore{
		db:        db.CacheWrap(),
		version:   rs.version,
		substores: make(map[string]cwwWriter), len(rs.substores),
	}
	for name, substore := range rs.substores {
		cms.substores[name] = substore.CacheWrap().(cwWriter)
	}
	return cms
}

func (cms cacheMultiStore) Write() {
	cms.db.Write()
	for substore := range rs.substores {
		substore.(cwWriter).Write()
	}
}
