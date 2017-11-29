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

	// Cache wrap MultiStore.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// Convenience
	GetStore(name string) interface{}
	GetKVStore(name string) KVStore
	GetIterKVStore(name string) IterKVStore
}

type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}

//----------------------------------------

// rootMultiStore is composed of many Committers.
// Implements MultiStore.
type rootMultiStore struct {
	db           dbm.DB
	version      int64
	storeLoaders map[string]CommitterLoader
	substores    map[string]Committer
}

func NewMultiStore(db dbm.DB) *rootMultiStore {
	return &rootMultiStore{
		db:           db,
		version:      0,
		storeLoaders: make(map[string]CommitterLoader),
		substores:    make(map[string]Committer),
	}
}

func (rs *rootMultiStore) SetCommitterLoader(name string, loader CommitterLoader) {
	if _, ok := rs.storeLoaders[name]; ok {
		panic(fmt.Sprintf("rootMultiStore duplicate substore name " + name))
	}
	rs.storeLoaders[name] = loader
}

//----------------------------------------
// rootMultiStore state

type msState struct {
	Substores []substore
}

func (ms *msState) Sort() {
	ms.Substores.Sort()
}

func (ms *msState) Hash() []byte {
	m := make(map[string]interface{}, len(ms.Substores))
	for _, substore := range ms.Substores {
		m[substore.name] = substore.subState
	}
	return merkle.SimpleHashFromMap(m)
}

//----------------------------------------
// substore state

type substore struct {
	name string
	subState
}

// This gets serialized by go-wire
type subState struct {
	CommitID CommitID
	// ... maybe add more state
}

func (ss subState) Hash() []byte {
	ssBytes, _ := wire.Marshal(ss) // Does not error
	hasher := ripemd160.New()
	hasher.Write(ssBytes)
	return hasher.Sum(nil)
}

//----------------------------------------

// Call once after all calls to SetCommitterLoader are complete.
func (rs *rootMultiStore) LoadLatestVersion() error {
	ver := rs.getLatestVersion()
	rs.LoadVersion(ver)
}

func (rs *rootMultiStore) getLatestVersion() int64 {
	var latest int64
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

// NOTE: Returns 0 unless LoadVersion() or LoadLatestVersion() is called.
func (rs *rootMultiStore) GetVersion() int64 {
	return rs.version
}

func (rs *rootMultiStore) LoadVersion(ver int64) error {

	// Special logic for version 0
	if ver == 0 {
		for name, storeLoader := range rs.storeLoaders {
			store, err := storeLoader(CommitID{Version: 0})
			if err != nil {
				return fmt.Errorf("Failed to load rootMultiStore: %v", err)
			}
			rs.substores[name] = store
		}
		return nil
	}
	// Otherwise, version is 1 or greater

	msStateKey := fmt.Sprintf(msStateKeyFmt, ver)
	stateBytes := rs.db.Get(msStateKey, ver)
	if bz == nil {
		return fmt.Errorf("Failed to load rootMultiStore: no data")
	}
	var state msState
	err := wire.Unmarshal(stateBytes, &state)
	if err != nil {
		return fmt.Errorf("Failed to load rootMultiStore: %v", err)
	}

	// Load each Substore
	var newSubstores = make(map[string]Committer)
	for _, store := range state.Substores {
		name, commitID := store.Name, store.CommitID
		storeLoader := rs.storeLoaders[name]
		if storeLoader == nil {
			return fmt.Errorf("Failed to loadrootMultiStore: CommitterLoader missing for %v", name)
		}
		store, err := storeLoader(commitID)
		if err != nil {
			return fmt.Errorf("Failed to load rootMultiStore: %v", err)
		}
		newSubstores[name] = store
	}

	// If any CommitterLoaders were not used, return error.
	for name := range rs.storeLoaders {
		if _, ok := rs.substores[name]; !ok {
			return fmt.Errorf("Unused CommitterLoader: %v", name)
		}
	}

	// Success.
	rs.version = ver
	rs.substores = newSubstores
	return nil
}

// Implements Committer
func (rs *rootMultiStore) Commit() CommitID {

	// Needs to be transactional
	batch := rs.db.NewBatch()

	// Save msState
	var state msState
	for name, store := range rs.substores {
		commitID := store.Commit()
		state.Substores = append(state.Substores,
			subState{
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

// Implements MultiStore
func (rs *rootMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStore(rs)
}

// Implements MultiStore
func (rs *rootMultiStore) GetCommitter(name string) Committer {
	return rs.store[name]
}

// Implements MultiStore
func (rs *rootMultiStore) GetKVStore(name string) KVStore {
	return rs.store[name].(KVStore)
}

// Implements MultiStore
func (rs *rootMultiStore) GetIterKVStore(name string) IterKVStore {
	return rs.store[name].(IterKVStore)
}

//----------------------------------------
// subStates

type subStates []subState

func (ssz subStates) Len() int           { return len(ssz) }
func (ssz subStates) Less(i, j int) bool { return ssz[i].Key < ssz[j].Key }
func (ssz subStates) Swap(i, j int)      { ssz[i], ssz[j] = ssz[j], ssz[i] }
func (ssz subStates) Sort()              { sort.Sort(ssz) }

func (ssz subStates) Hash() []byte {
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
	version   int64
	substores map[string]cwWriter
}

func newCacheMultiStore(rs *rootMultiStore) cacheMultiStore {
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

// Implements CacheMultiStore
func (cms cacheMultiStore) Write() {
	cms.db.Write()
	for substore := range rs.substores {
		substore.(cwWriter).Write()
	}
}

// Implements CacheMultiStore
func (rs cacheMultiStore) CacheMultiStore() CacheMultiStore {
	return newCacheMultiStore(rs)
}

// Implements CacheMultiStore
func (rs cacheMultiStore) GetCommitter(name string) Committer {
	return rs.store[name]
}

// Implements CacheMultiStore
func (rs cacheMultiStore) GetKVStore(name string) KVStore {
	return rs.store[name].(KVStore)
}

// Implements CacheMultiStore
func (rs cacheMultiStore) GetIterKVStore(name string) IterKVStore {
	return rs.store[name].(IterKVStore)
}
