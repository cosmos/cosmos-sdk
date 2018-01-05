package store

import (
	"sync"

	"github.com/tendermint/iavl"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

// iavlStoreLoader contains info on what store we want to load from
type iavlStoreLoader struct {
	db         dbm.DB
	cacheSize  int
	numHistory int64
}

// NewIAVLStoreLoader returns a CommitStoreLoader that returns an iavlStore
func NewIAVLStoreLoader(db dbm.DB, cacheSize int, numHistory int64) CommitStoreLoader {
	l := iavlStoreLoader{
		db:         db,
		cacheSize:  cacheSize,
		numHistory: numHistory,
	}
	return l.Load
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

var _ KVStore = (*iavlStore)(nil)
var _ CommitStore = (*iavlStore)(nil)

// iavlStore Implements KVStore and CommitStore.
type iavlStore struct {

	// The underlying tree.
	tree *iavl.VersionedTree

	// How many old versions we hold onto.
	// A value of 0 means keep all history.
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
	if st.numHistory > 0 && (st.numHistory < st.tree.Version64()) {
		toRelease := version - st.numHistory
		st.tree.DeleteVersion(toRelease)
	}

	return CommitID{
		Version: version,
		Hash:    hash,
	}
}

// CacheWrap implements KVStore.
func (st *iavlStore) CacheWrap() CacheWrap {
	return NewCacheKVStore(st)
}

// Set implements KVStore.
func (st *iavlStore) Set(key, value []byte) {
	st.tree.Set(key, value)
}

// Get implements KVStore.
func (st *iavlStore) Get(key []byte) (value []byte) {
	_, v := st.tree.Get(key)
	return v
}

// Has implements KVStore.
func (st *iavlStore) Has(key []byte) (exists bool) {
	return st.tree.Has(key)
}

// Delete implements KVStore.
func (st *iavlStore) Delete(key []byte) {
	st.tree.Remove(key)
}

// Iterator implements KVStore.
func (st *iavlStore) Iterator(start, end []byte) Iterator {
	return newIAVLIterator(st.tree.Tree(), start, end, true)
}

// ReverseIterator implements IterKVStore.
func (st *iavlStore) ReverseIterator(start, end []byte) Iterator {
	return newIAVLIterator(st.tree.Tree(), start, end, false)
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
	iterCh chan cmn.KVPair

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
		iterCh:    make(chan cmn.KVPair, 0), // Set capacity > 0?
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
			case iter.iterCh <- cmn.KVPair{key, value}:
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

// Close implements Iterator
func (iter *iavlIterator) Close() {
	close(iter.quitCh)
}

//----------------------------------------

func (iter *iavlIterator) setNext(key, value []byte) {
	iter.assertIsValid()

	iter.key = key
	iter.value = value
}

func (iter *iavlIterator) setInvalid() {
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

func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}
