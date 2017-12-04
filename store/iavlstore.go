// XXX Need to s/Committer/CommitStore/g

package store

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tmlibs/db"
)

// NewIAVLLoader returns a CommitterLoader that returns
// an IAVLCommitter
func NewIAVLLoader(dbName string, cacheSize int, nHistoricalVersions uint64) CommitterLoader {
	l := iavlLoader{
		dbName:              dbName,
		cacheSize:           cacheSize,
		nHistoricalVersions: nHistoricalVersions,
	}
	return CommitterLoader(l.Load)
}

var _ CacheIterKVStore = (*IAVLCommitter)(nil)
var _ Committer = (*IAVLCommitter)(nil)

// IAVLCommitter Implements IterKVStore and Committer
type IAVLCommitter struct {
	// we must store the last height here, as it is needed
	// for saving the versioned tree
	lastHeight uint64

	// nHistoricalVersions is how many old versions we hold onto,
	// uses a naive "hold last X versions" algorithm
	nHistoricalVersions uint64

	// this is all historical data and connection to
	// the db
	tree *iavl.VersionedTree

	// this is the current working state to be saved
	// on the next commit
	IAVLStore
}

// NewIAVLCommitter properly initializes a committer
// that is ready to use as a IterKVStore
func NewIAVLCommitter(tree *iavl.VersionedTree,
	lastHeight uint64, nHistoricalVersions uint64) *IAVLCommitter {
	ic := &IAVLCommitter{
		tree:                tree,
		lastHeight:          lastHeight,
		nHistoricalVersions: nHistoricalVersions,
	}
	ic.updateStore()
	return ic
}

// Commit syncs the working state and
// saves another version to the db
func (i *IAVLCommitter) Commit() CommitID {
	// TODO: sync working state??
	// I think this is done already just by writing to tree.Tree()

	// save a new version
	ic.lastHeight++
	hash, err := ic.tree.SaveVersion(ic.lastHeight)
	if err != nil {
		// TODO: do we want to extend Commit to
		// allow returning errors?
		panic(err)
	}

	// now point working state to the new status
	ic.updateStore()

	// release an old version of history
	if ic.nHistoricalVersions <= ic.lastHeight {
		release := ic.lastHeight - ic.nHistoricalVersions
		ic.tree.DeleteVersion(release)
	}

	return CommitID{
		Version: ic.lastHeight,
		Hash:    hash,
	}
}

// store returns a wrapper around the current writable state
func (ic *IAVLCommitter) updateStore() {
	ic.IAVLStore = IAVLStore{ic.tree.Tree()}
}

// IAVLStore is the writable state (not history) and
// implements the IterKVStore interface.
type IAVLStore struct {
	tree *iavl.Tree
}

// CacheWrap implements IterKVStore.
func (is IAVLStore) CacheWrap() CacheWriter {
	return is.CacheIterKVStore()
}

// CacheIterKVStore implements IterKVStore.
func (is IAVLStore) CacheIterKVStore() CacheIterKVStore {
	// TODO: Add CacheWrap to IAVLTree.
	return i
}

// Set implements IterKVStore.
func (is IAVLStore) Set(key, value []byte) (prev []byte) {
	_, prev = is.tree.Get(key)
	is.tree.Set(key, value)
	return prev
}

// Get implements IterKVStore.
func (is IAVLStore) Get(key []byte) (value []byte, exists bool) {
	_, v := is.tree.Get(key)
	return v, (v != nil)
}

// Has implements IterKVStore.
func (is IAVLStore) Has(key []byte) (exists bool) {
	return is.tree.Has(key)
}

// Remove implements IterKVStore.
func (is IAVLStore) Remove(key []byte) (prev []byte, removed bool) {
	return is.tree.Remove(key)
}

// Iterator implements IterKVStore.
func (is IAVLStore) Iterator(start, end []byte) Iterator {
	// TODO: this needs changes to IAVL tree
	return nil
}

// ReverseIterator implements IterKVStore.
func (is IAVLStore) ReverseIterator(start, end []byte) Iterator {
	// TODO
	return nil
}

// First implements IterKVStore.
func (is IAVLStore) First(start, end []byte) (kv KVPair, ok bool) {
	// TODO
	return KVPair{}, false
}

// Last implements IterKVStore.
func (is IAVLStore) Last(start, end []byte) (kv KVPair, ok bool) {
	// TODO
	return KVPair{}, false
}

var _ IterKVStore = IAVLStore{}

type iavlIterator struct {
	// TODO
}

var _ Iterator = (*iavlIterator)(nil)

// Domain implements Iterator
//
// The start & end (exclusive) limits to iterate over.
// If end < start, then the Iterator goes in reverse order.
func (ii *iavlIterator) Domain() (start, end []byte) {
	// TODO
	return nil, nil
}

// Valid implements Iterator
//
// Returns if the current position is valid.
func (ii *iavlIterator) Valid() bool {
	// TODO
	return false
}

// Next implements Iterator
//
// Next moves the iterator to the next key/value pair.
func (ii *iavlIterator) Next() {
	// TODO
}

// Key implements Iterator
//
// Key returns the key of the current key/value pair, or nil if done.
// The caller should not modify the contents of the returned slice, and
// its contents may change after calling Next().
func (ii *iavlIterator) Key() []byte {
	// TODO
	return nil
}

// Value implements Iterator
//
// Value returns the key of the current key/value pair, or nil if done.
// The caller should not modify the contents of the returned slice, and
// its contents may change after calling Next().
func (ii *iavlIterator) Value() []byte {
	// TODO
	return nil
}

// Release implements Iterator
//
// Releases any resources and iteration-locks
func (ii *iavlIterator) Release() {
	// TODO
}

// iavlLoader contains info on what store we want to load from
type iavlLoader struct {
	dbName             string
	cacheSize          int
	nHistoricalVersion uint64
}

// Load implements CommitLoader type
func (il iavlLoader) Load(id CommitID) (Committer, error) {
	// memory backed case, just for testing
	if il.dbName == "" {
		tree := iavl.NewVersionedTree(0, dbm.NewMemDB())
		store := NewIAVLCommitter(tree, 0, il.nHistoricalVersions)
		return store, nil
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(il.dbName)
	if err != nil {
		return nil, errors.New("Invalid Database Name")
	}

	// Some external calls accidently add a ".db", which is now removed
	dbPath = strings.TrimSuffix(dbPath, path.Ext(dbPath))

	// Split the database name into it's components (dir, name)
	dir := filepath.Dir(dbPath)
	name := filepath.Base(dbPath)

	// Open database called "dir/name.db", if it doesn't exist it will be created
	db := dbm.NewDB(name, dbm.LevelDBBackendStr, dir)
	tree := iavl.NewVersionedTree(il.cacheSize, db)
	if err = tree.Load(); err != nil {
		return nil, errors.New("Loading tree: " + err.Error())
	}

	// TODO: load the version stored in id
	store := NewIAVLCommitter(tree, tree.LatestVersion(),
		il.nHistoricalVersions)
	return store, nil
}
