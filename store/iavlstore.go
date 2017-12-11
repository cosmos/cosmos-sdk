package store

import (
	"sync"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tmlibs/db"
)

// NewIAVLStoreLoader returns a CommitStoreLoader that returns an iavlStore
func NewIAVLStoreLoader(db dbm.DB, cacheSize int, numHistory int64) CommitStoreLoader {
	l := iavlStoreLoader{
		db:         db,
		cacheSize:  cacheSize,
		numHistory: numHistory,
	}
	return l.Load
}

var _ IterKVStore = (*iavlStore)(nil)
var _ CommitStore = (*iavlStore)(nil)

// iavlStore Implements IterKVStore and CommitStore.
type iavlStore struct {

	// The underlying tree.
	tree *iavl.VersionedTree

	// How many old versions we hold onto.
	numHistory int64
}

// CONTRACT: tree should be fully loaded.
func newIAVLStore(tree *iavl.VersionedTree, numHistory int64) *iavlStore {
	st := &iavlStore{
		tree:       tree,
		numHistory: numHistory,
	}
	return st
}

// Commit persists the store.
func (st *iavlStore) Commit() CommitID {

	// Save a new version.
	hash, version, err := st.tree.SaveVersion()
	if err != nil {
		// TODO: Do we want to extend Commit to allow returning errors?
		panic(err)
	}

	// Release an old version of history
	if st.numHistory < st.tree.Version64() {
		toRelease := version - st.numHistory
		st.tree.DeleteVersion(toRelease)
	}

	return CommitID{
		Version: version,
		Hash:    hash,
	}
}

// CacheWrap implements IterKVStore.
func (st *iavlStore) CacheWrap() CacheWrap {
	return st.CacheIterKVStore()
}

// CacheKVStore implements IterKVStore.
func (st *iavlStore) CacheKVStore() CacheKVStore {
	return st.CacheIterKVStore()
}

// CacheIterKVStore implements IterKVStore.
func (st *iavlStore) CacheIterKVStore() CacheIterKVStore {
	// XXX Create generic IterKVStore wrapper.
	return nil
}

// Set implements IterKVStore.
func (st *iavlStore) Set(key, value []byte) {
	st.tree.Set(key, value)
}

// Get implements IterKVStore.
func (st *iavlStore) Get(key []byte) (value []byte) {
	_, v := st.tree.Get(key)
	return v
}

// Has implements IterKVStore.
func (st *iavlStore) Has(key []byte) (exists bool) {
	return st.tree.Has(key)
}

// Remove implements IterKVStore.
func (st *iavlStore) Remove(key []byte) {
	st.tree.Remove(key)
}

// Iterator implements IterKVStore.
func (st *iavlStore) Iterator(start, end []byte) Iterator {
	return newIAVLIterator(st.tree.Tree(), start, end, true)
}

// ReverseIterator implements IterKVStore.
func (st *iavlStore) ReverseIterator(start, end []byte) Iterator {
	return newIAVLIterator(st.tree.Tree(), start, end, false)
}

// First implements IterKVStore.
func (st *iavlStore) First(start, end []byte) (kv KVPair, ok bool) {
	return iteratorFirst(st, start, end)
}

// Last implements IterKVStore.
func (st *iavlStore) Last(start, end []byte) (kv KVPair, ok bool) {
	return iteratorLast(st, start, end)
}

//----------------------------------------

// Implements Iterator
type iavlIterator struct {
	// Underlying store
	tree *iavl.Tree

	// Domain
	start, end []byte

	// Iteration order
	ascending bool

	// Channel to push iteration values.
	iterCh chan KVPair

	// Close this to release goroutine.
	quitCh chan struct{}

	// Close this to signal that state is initialized.
	initCh chan struct{}

	//----------------------------------------
	// What follows are mutable state.
	mtx sync.Mutex

	invalid bool   // True once, true forever
	key     []byte // The current key
	value   []byte // The current value
}

var _ Iterator = (*iavlIterator)(nil)

// newIAVLIterator will create a new iavlIterator.
// CONTRACT: Caller must release the iavlIterator, as each one creates a new
// goroutine.
func newIAVLIterator(tree *iavl.Tree, start, end []byte, ascending bool) *iavlIterator {
	iter := &iavlIterator{
		tree:      tree,
		start:     cp(start),
		end:       cp(end),
		ascending: ascending,
		iterCh:    make(chan KVPair, 0), // Set capacity > 0?
		quitCh:    make(chan struct{}),
		initCh:    make(chan struct{}),
	}
	go iter.iterateRoutine()
	go iter.initRoutine()
	return iter
}

// Run this to funnel items from the tree to iterCh.
func (iter *iavlIterator) iterateRoutine() {
	iter.tree.IterateRange(
		iter.start, iter.end, iter.ascending,
		func(key, value []byte) bool {
			select {
			case <-iter.quitCh:
				return true // done with iteration.
			case iter.iterCh <- KVPair{key, value}:
				return false // yay.
			}
		},
	)
	close(iter.iterCh) // done.
}

// Run this to fetch the first item.
func (iter *iavlIterator) initRoutine() {
	iter.receiveNext()
	close(iter.initCh)
}

// Domain implements Iterator
func (iter *iavlIterator) Domain() (start, end []byte) {
	return iter.start, iter.end
}

// Valid implements Iterator
func (iter *iavlIterator) Valid() bool {
	iter.waitInit()
	iter.mtx.Lock()
	defer iter.mtx.Unlock()

	return !iter.invalid
}

// Next implements Iterator
func (iter *iavlIterator) Next() {
	iter.waitInit()
	iter.mtx.Lock()
	defer iter.mtx.Unlock()
	iter.assertIsValid()

	iter.receiveNext()
}

// Key implements Iterator
func (iter *iavlIterator) Key() []byte {
	iter.waitInit()
	iter.mtx.Lock()
	defer iter.mtx.Unlock()
	iter.assertIsValid()

	return iter.key
}

// Value implements Iterator
func (iter *iavlIterator) Value() []byte {
	iter.waitInit()
	iter.mtx.Lock()
	defer iter.mtx.Unlock()
	iter.assertIsValid()

	return iter.value
}

// Release implements Iterator
func (iter *iavlIterator) Release() {
	close(iter.quitCh)
}

//----------------------------------------

func (iter *iavlIterator) setNext(key, value []byte) {
	iter.mtx.Lock()
	defer iter.mtx.Unlock()
	iter.assertIsValid()

	iter.key = key
	iter.value = value
}

func (iter *iavlIterator) setInvalid() {
	iter.mtx.Lock()
	defer iter.mtx.Unlock()
	iter.assertIsValid()

	iter.invalid = true
}

func (iter *iavlIterator) waitInit() {
	<-iter.initCh
}

func (iter *iavlIterator) receiveNext() {
	kvPair, ok := <-iter.iterCh
	if ok {
		iter.setNext(kvPair.Key, kvPair.Value)
	} else {
		iter.setInvalid()
	}
}

func (iter *iavlIterator) assertIsValid() {
	if iter.invalid {
		panic("invalid iterator")
	}
}

//----------------------------------------

// iavlStoreLoader contains info on what store we want to load from
type iavlStoreLoader struct {
	db         dbm.DB
	cacheSize  int
	numHistory int64
}

// Load implements CommitLoader.
func (isl iavlStoreLoader) Load(id CommitID) (CommitStore, error) {
	tree := iavl.NewVersionedTree(isl.db, isl.cacheSize)
	err := tree.Load()
	if err != nil {
		return nil, err
	}
	store := newIAVLStore(tree, isl.numHistory)
	return store, nil
}

//----------------------------------------

func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}
