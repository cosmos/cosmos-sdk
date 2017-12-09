package store

import (
	"github.com/tendermint/go-wire/data"
)

type CommitID struct {
	Version int64
	Hash    []byte
}

func (cid CommitID) IsZero() bool {
	return cid.Version == 0 && len(cid.Hash) == 0
}

type Committer interface {
	// Commit persists the state to disk.
	Commit() CommitID
}

type CacheWrapper interface {
	/*
		CacheWrap() makes the most appropriate cache-wrap.  For example,
		IAVLStore.CacheWrap() returns a CacheIterKVStore.  After call to
		.Write() on the cache-wrap, all previous cache-wraps on the object
		expire.

		CacheWrap() should not return a Committer, since Commit() on
		cache-wraps make no sense.  It can return KVStore, IterKVStore, etc.

		The returned object may or may not implement CacheWrap() as well.

		NOTE: https://dave.cheney.net/2017/07/22/should-go-2-0-support-generics.
	*/
	CacheWrap() CacheWrap
}

type CacheWrap interface {
	// Write syncs with the underlying store.
	Write()

	// CacheWrap recursively wraps again.
	CacheWrap() CacheWrap
}

type CommitStore interface {
	Committer
	CacheWrapper
}

type CommitStoreLoader func(id CommitID) (CommitStore, error)

// KVStore is a simple interface to get/set data
type KVStore interface {
	Set(key, value []byte) (prev []byte)
	Get(key []byte) (value []byte, exists bool)
	Has(key []byte) (exists bool)
	Remove(key []byte) (prev []byte, removed bool)

	// CacheKVStore() wraps a thing with a cache.  After
	// calling .Write() on the CacheKVStore, all previous
	// cache-wraps on the object expire.
	CacheKVStore() CacheKVStore

	// CacheWrap() returns a CacheKVStore.
	CacheWrap() CacheWrap
}

type CacheKVStore interface {
	KVStore
	Write() // Writes operations to underlying KVStore
}

// IterKVStore can be iterated on
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
type IterKVStore interface {
	KVStore

	Iterator(start, end []byte) Iterator
	ReverseIterator(start, end []byte) Iterator

	First(start, end []byte) (kv KVPair, ok bool)
	Last(start, end []byte) (kv KVPair, ok bool)

	// CacheIterKVStore() wraps a thing with a cache.
	// After calling .Write() on the CacheIterKVStore, all
	// previous cache-wraps on the object expire.
	CacheIterKVStore() CacheIterKVStore

	// CacheWrap() returns a CacheIterKVStore.
	// CacheWrap() defined in KVStore
}

type CacheIterKVStore interface {
	IterKVStore
	Write() // Writes operations to underlying KVStore
}

type KVPair struct {
	Key   data.Bytes
	Value data.Bytes
}

/*
	Usage:

	for itr := kvm.Iterator(start, end); itr.Valid(); itr.Next() {
		k, v := itr.Key(); itr.Value()
		....
	}
*/
type Iterator interface {

	// The start & end (exclusive) limits to iterate over.
	// If end < start, then the Iterator goes in reverse order.
	// A domain of ([]byte{12, 13}, []byte{12, 14}) will iterate
	// over anything with the prefix []byte{12, 13}
	Domain() (start []byte, end []byte)

	// Returns if the current position is valid.
	Valid() bool

	// Next moves the iterator to the next key/value pair.
	//
	// If Valid returns false, this method will panic.
	Next()

	// Key returns the key of the current key/value pair, or nil if done.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change after calling Next().
	//
	// If Valid returns false, this method will panic.
	Key() []byte

	// Value returns the key of the current key/value pair, or nil if done.
	// The caller should not modify the contents of the returned slice, and
	// its contents may change after calling Next().
	//
	// If Valid returns false, this method will panic.
	Value() []byte

	// Releases any resources and iteration-locks
	Release()
}

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

	// CacheWrap returns a CacheMultiStore.
	CacheWrap() CacheWrap

	// Convenience
	GetStore(name string) interface{}
	GetKVStore(name string) KVStore
	GetIterKVStore(name string) IterKVStore
}

type CacheMultiStore interface {
	MultiStore
	Write() // Writes operations to underlying KVStore
}
