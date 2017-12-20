package store

import (
	"bytes"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/db"
)

//----------------------------------------
// MultiStore

type MultiStore interface {

	// Last commit, or the zero CommitID.
	// If not zero, CommitID.Version is CurrentVersion()-1.
	LastCommitID() CommitID

	// Current version being worked on now, not yet committed.
	// Should be greater than 0.
	CurrentVersion() int64

	// Cache wrap MultiStore.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// Convenience
	GetStore(name string) interface{}
	GetKVStore(name string) KVStore
}

type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}

type CommitStore interface {
	Committer
	CacheWrapper
}

type CommitStoreLoader func(id CommitID) (CommitStore, error)

type Committer interface {
	// Commit persists the state to disk.
	Commit() CommitID
}

//----------------------------------------
// KVStore

// KVStore is a simple interface to get/set data
type KVStore interface {

	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be greater than end, or the Iterator is invalid.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	ReverseIterator(start, end []byte) Iterator
}

// db.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = db.DB(nil)

// Alias iterator to db's Iterator for convenience.
type Iterator = db.Iterator

// CacheKVStore cache-wraps a KVStore.  After calling .Write() on the
// CacheKVStore, all previously created CacheKVStores on the object expire.
type CacheKVStore interface {
	KVStore
	Write() // Writes operations to underlying KVStore
}

//----------------------------------------
// CacheWrap

/*
	CacheWrap() makes the most appropriate cache-wrap.  For example,
	IAVLStore.CacheWrap() returns a CacheKVStore.

	CacheWrap() should not return a Committer, since Commit() on
	cache-wraps make no sense.  It can return KVStore, HeapStore,
	SpaceStore, etc.
*/
type CacheWrapper interface {
	CacheWrap() CacheWrap
}

type CacheWrap interface {

	// Write syncs with the underlying store.
	Write()

	// CacheWrap recursively wraps again.
	CacheWrap() CacheWrap
}

//----------------------------------------
// etc

type KVPair struct {
	Key   data.Bytes
	Value data.Bytes
}

// CommitID contains the tree version number and its merkle root.
type CommitID struct {
	Version int64
	Hash    []byte
}

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}

// bytes.Compare but bounded on both sides by nil.
// both (k1, nil) and (nil, k2) return -1
func keyCompare(k1, k2 []byte) int {
	if k1 == nil && k2 == nil {
		return 0
	} else if k1 == nil || k2 == nil {
		return -1
	}
	return bytes.Compare(k1, k2)
}
