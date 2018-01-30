package types

import (
	"fmt"

	dbm "github.com/tendermint/tmlibs/db"
)

// NOTE: These are implemented in cosmos-sdk/store.

type Store interface {
	GetStoreType() StoreType
	CacheWrapper
}

// Something that can persist to disk.
type Committer interface {
	Commit() CommitID
	LastCommitID() CommitID
}

// Stores of MultiStore must implement CommitStore.
type CommitStore interface {
	Committer
	Store
}

//----------------------------------------
// MultiStore

type MultiStore interface {
	Store

	// Cache wrap MultiStore.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// Convenience for fetching substores.
	GetStore(StoreKey) Store
	GetKVStore(StoreKey) KVStore

	GetStoreByName(string) Store
}

// From MultiStore.CacheMultiStore()....
type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}

// A non-cache MultiStore.
type CommitMultiStore interface {
	Committer
	MultiStore

	// Mount a store of type.
	MountStoreWithDB(key StoreKey, typ StoreType, db dbm.DB)

	// Panics on a nil key.
	GetCommitStore(key StoreKey) CommitStore

	// Load the latest persisted version.  Called once after all
	// calls to Mount*Store() are complete.
	LoadLatestVersion() error

	// Load a specific persisted version.  When you load an old
	// version, or when the last commit attempt didn't complete,
	// the next commit after loading must be idempotent (return the
	// same commit id).  Otherwise the behavior is undefined.
	LoadVersion(ver int64) error
}

//----------------------------------------
// KVStore

// KVStore is a simple interface to get/set data
type KVStore interface {
	Store

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

	// TODO Not yet implemented.
	// CreateSubKVStore(key *storeKey) (KVStore, error)

	// TODO Not yet implemented.
	// GetSubKVStore(key *storeKey) KVStore

}

// Alias iterator to db's Iterator for convenience.
type Iterator = dbm.Iterator

// CacheKVStore cache-wraps a KVStore.  After calling .Write() on
// the CacheKVStore, all previously created CacheKVStores on the
// object expire.
type CacheKVStore interface {
	KVStore

	// Writes operations to underlying KVStore
	Write()
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
// CommitID

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

//----------------------------------------
// Store types

type StoreType int

const (
	StoreTypeMulti StoreType = iota
	StoreTypeDB
	StoreTypeIAVL
)

//----------------------------------------
// Keys for accessing substores

// StoreKey is a key used to index stores in a MultiStore.
type StoreKey interface {
	Name() string
	String() string
}

// KVStoreKey is used for accessing substores.
// Only the pointer value should ever be used - it functions as a capabilities key.
type KVStoreKey struct {
	name string
}

// NewKVStoreKey returns a new pointer to a KVStoreKey.
// Use a pointer so keys don't collide.
func NewKVStoreKey(name string) *KVStoreKey {
	return &KVStoreKey{
		name: name,
	}
}

func (key *KVStoreKey) Name() string {
	return key.name
}

func (key *KVStoreKey) String() string {
	return fmt.Sprintf("KVStoreKey{%p, %s}", key, key.name)
}
