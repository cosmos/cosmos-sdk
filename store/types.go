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

type CommitterLoader func(id CommitID) (Committer, error)

// KVStore is a simple interface to get/set data
type KVStore interface {
	Set(key, value []byte) (prev []byte)
	Get(key []byte) (value []byte, exists bool)
	Has(key []byte) (exists bool)
	Remove(key []byte) (prev []byte, removed bool)

	// CacheKVStore() wraps a thing with a cache.  After
	// calling .Write() on the CacheKVStore, all previous
	// CacheWraps on the object expire.
	CacheKVStore() CacheKVStore
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
	// previous CacheWraps on the object expire.
	CacheIterKVStore() CacheIterKVStore
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
