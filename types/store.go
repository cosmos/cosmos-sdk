package types

import (
	"fmt"

	dbm "github.com/tendermint/tmlibs/db"
)

// NOTE: These are implemented in cosmos-sdk/store.

//----------------------------------------
// MultiStore

type MultiStore interface {

	// Last commit, or the zero CommitID.
	// If not zero, CommitID.Version is NextVersion()-1.
	LastCommitID() CommitID

	// Current version being worked on now, not yet committed.
	// Should be greater than 0.
	NextVersion() int64

	// Cache wrap MultiStore.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// Convenience
	GetStore(name string) interface{}
	GetKVStore(name string) KVStore
}

// From MultiStore.CacheMultiStore()....
type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}

// Substores of MultiStore must implement CommitStore.
type CommitStore interface {
	Committer
	CacheWrapper
}

// A non-cache store that can commit (persist) and get a Merkle root.
type Committer interface {
	Commit() CommitID
}

// A non-cache MultiStore.
type CommitMultiStore interface {
	CommitStore
	MultiStore

	// Add a substore loader.
	SetSubstoreLoader(name string, loader CommitStoreLoader)

	// Load the latest persisted version.
	LoadLatestVersion() error

	// Load a specific persisted version.  When you load an old version, or
	// when the last commit attempt didn't complete, the next commit after
	// loading must be idempotent (return the same commit id).  Otherwise the
	// behavior is undefined.
	LoadVersion(ver int64) error
}

// These must be added to the MultiStore before calling LoadVersion() or
// LoadLatest().
type CommitStoreLoader func(id CommitID) (CommitStore, error)

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

// dbm.DB implements KVStore so we can CacheKVStore it.
var _ KVStore = dbm.DB(nil)

// Alias iterator to db's Iterator for convenience.
type Iterator = dbm.Iterator

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

// CommitID contains the tree version number and its merkle root.
type CommitID struct {
	Version int64
	Hash    []byte
}

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}

func (cid CommitID) String() string {
	return fmt.Sprintf("CommitID{%v:%X}", cid.Hash, cid.Version)
}
