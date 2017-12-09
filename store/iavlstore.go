package store

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tmlibs/db"
)

// NewIAVLStoreLoader returns a CommitStoreLoader that returns an iavlStore
func NewIAVLStoreLoader(dbName string, cacheSize int, numHistory int64) CommitStoreLoader {
	l := iavlStoreLoader{
		dbName:     dbName,
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
	if st.numHistory < st.tree.Version() {
		toRelease := version - st.numHistory
		st.tree.DeleteVersion(toRelease)
	}

	return CommitID{
		Version: version,
		Hash:    hash,
	}
}

// CacheWrap implements IterKVStore.
func (st *iavlStore) CacheWrap() CacheWriter {
	return st.CacheIterKVStore()
}

// CacheIterKVStore implements IterKVStore.
func (st *iavlStore) CacheIterKVStore() CacheIterKVStore {
	// XXX Create generic IterKVStore wrapper.
	return nil
}

// Set implements IterKVStore.
func (st *iavlStore) Set(key, value []byte) (prev []byte) {
	_, prev = st.tree.Get(key)
	st.tree.Set(key, value)
	return prev
}

// Get implements IterKVStore.
func (st *iavlStore) Get(key []byte) (value []byte, exists bool) {
	_, v := st.tree.Get(key)
	return v, (v != nil)
}

// Has implements IterKVStore.
func (st *iavlStore) Has(key []byte) (exists bool) {
	return st.tree.Has(key)
}

// Remove implements IterKVStore.
func (st *iavlStore) Remove(key []byte) (prev []byte, removed bool) {
	return st.tree.Remove(key)
}

// Iterator implements IterKVStore.
func (st *iavlStore) Iterator(start, end []byte) Iterator {
	// XXX Create iavlIterator (without modifying tendermint/iavl)
	return nil
}

// ReverseIterator implements IterKVStore.
func (st *iavlStore) ReverseIterator(start, end []byte) Iterator {
	// XXX Create iavlIterator (without modifying tendermint/iavl)
	return nil
}

// First implements IterKVStore.
func (is IAVLStore) First(start, end []byte) (kv KVPair, ok bool) {
	// XXX
	return KVPair{}, false
}

// Last implements IterKVStore.
func (is IAVLStore) Last(start, end []byte) (kv KVPair, ok bool) {
	// XXX
	return KVPair{}, false
}

//----------------------------------------

type iavlIterator struct {
	// TODO
}

var _ Iterator = (*iavlIterator)(nil)

// Domain implements Iterator
func (ii *iavlIterator) Domain() (start, end []byte) {
	// TODO
	return nil, nil
}

// Valid implements Iterator
func (ii *iavlIterator) Valid() bool {
	// TODO
	return false
}

// Next implements Iterator
func (ii *iavlIterator) Next() {
	// TODO
}

// Key implements Iterator
func (ii *iavlIterator) Key() []byte {
	// TODO
	return nil
}

// Value implements Iterator
func (ii *iavlIterator) Value() []byte {
	// TODO
	return nil
}

// Release implements Iterator
func (ii *iavlIterator) Release() {
	// TODO
}

//----------------------------------------

// iavlStoreLoader contains info on what store we want to load from
type iavlStoreLoader struct {
	db         dbm.DB
	cacheSize  int
	numHistory int64
}

// Load implements CommitLoader.
func (isl iavlLoader) Load(id CommitID) (CommitStore, error) {
	tree := iavl.NewVersionedTree(isl.db, isl.cacheSize)
	err := tree.Load()
	if err != nil {
		return nil, err
	}
	store := newIAVLStore(tree, isl.numHistory)
	return store, nil
}
