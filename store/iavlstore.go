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

var _ CacheWrappable = (*IAVLCommitter)(nil)
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
	i := &IAVLCommitter{
		tree:                tree,
		lastHeight:          lastHeight,
		nHistoricalVersions: nHistoricalVersions,
	}
	i.updateStore()
	return i
}

// Commit syncs the working state and
// saves another version to the db
func (i *IAVLCommitter) Commit() CommitID {
	// TODO: sync working state??
	// I think this is done already just by writing to tree.Tree()

	// save a new version
	i.lastHeight++
	hash, err := i.tree.SaveVersion(i.lastHeight)
	if err != nil {
		// TODO: do we want to extend Commit to
		// allow returning errors?
		panic(err)
	}

	// now point working state to the new status
	i.updateStore()

	// release an old version of history
	if i.nHistoricalVersions <= i.lastHeight {
		release := i.lastHeight - i.nHistoricalVersions
		i.tree.DeleteVersion(release)
	}

	return CommitID{
		Version: i.lastHeight,
		Hash:    hash,
	}
}

// store returns a wrapper around the current writable state
func (i *IAVLCommitter) updateStore() {
	i.IAVLStore = IAVLStore{i.tree.Tree()}
}

// IAVLStore is the writable state (not history) and
// implements the IterKVStore interface.
type IAVLStore struct {
	tree *iavl.Tree
}

// CacheWrap returns a wrapper around the current writable state
func (i IAVLStore) CacheWrap() CacheWriter {
	// TODO: add CacheWrap to IAVLTree using new db stuff
	return i
}

// Set implements KVStore
func (i IAVLStore) Set(key, value []byte) (prev []byte) {
	_, prev = i.tree.Get(key)
	i.tree.Set(key, value)
	return prev
}

// Get implements KVStore
func (i IAVLStore) Get(key []byte) (value []byte, exists bool) {
	_, v := i.tree.Get(key)
	return v, (v != nil)
}

// Has implements KVStore
func (i IAVLStore) Has(key []byte) (exists bool) {
	return i.tree.Has(key)
}

// Remove implements KVStore
func (i IAVLStore) Remove(key []byte) (prev []byte, removed bool) {
	return i.tree.Remove(key)
}

// Iterator implements IterKVStore
func (i IAVLStore) Iterator(start, end []byte) Iterator {
	// TODO: this needs changes to IAVL tree
	return nil
}

// ReverseIterator implements IterKVStore
func (i IAVLStore) ReverseIterator(start, end []byte) Iterator {
	// TODO
	return nil
}

// First implements IterKVStore
func (i IAVLStore) First(start, end []byte) (kv KVPair, ok bool) {
	// TODO
	return KVPair{}, false
}

// Last implements IterKVStore
func (i IAVLStore) Last(start, end []byte) (kv KVPair, ok bool) {
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
func (i *iavlIterator) Domain() (start, end []byte) {
	// TODO
	return nil, nil
}

// Valid implements Iterator
//
// Returns if the current position is valid.
func (i *iavlIterator) Valid() bool {
	// TODO
	return false
}

// Next implements Iterator
//
// Next moves the iterator to the next key/value pair.
func (i *iavlIterator) Next() {
	// TODO
}

// Key implements Iterator
//
// Key returns the key of the current key/value pair, or nil if done.
// The caller should not modify the contents of the returned slice, and
// its contents may change after calling Next().
func (i *iavlIterator) Key() []byte {
	// TODO
	return nil
}

// Value implements Iterator
//
// Value returns the key of the current key/value pair, or nil if done.
// The caller should not modify the contents of the returned slice, and
// its contents may change after calling Next().
func (i *iavlIterator) Value() []byte {
	// TODO
	return nil
}

// Release implements Iterator
//
// Releases any resources and iteration-locks
func (i *iavlIterator) Release() {
	// TODO
}

// iavlLoader contains info on what store we want to load from
type iavlLoader struct {
	dbName             string
	cacheSize          int
	nHistoricalVersion uint64
}

// Load implements CommitLoader type
func (l iavlLoader) Load(id CommitID) (Committer, error) {
	// memory backed case, just for testing
	if l.dbName == "" {
		tree := iavl.NewVersionedTree(0, dbm.NewMemDB())
		store := NewIAVLCommitter(tree, 0, l.nHistoricalVersions)
		return store, nil
	}

	// Expand the path fully
	dbPath, err := filepath.Abs(l.dbName)
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
	tree := iavl.NewVersionedTree(l.cacheSize, db)
	if err = tree.Load(); err != nil {
		return nil, errors.New("Loading tree: " + err.Error())
	}

	// TODO: load the version stored in id
	store := NewIAVLCommitter(tree, tree.LatestVersion(),
		l.nHistoricalVersions)
	return store, nil
}
