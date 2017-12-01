package store

import (
	"github.com/tendermint/go-wire/data"
)

type CommitID struct {
	Version int64
	Hash    []byte
}

type (cid CommitID) IsZero() bool {
	return cid.Version == 0  && len(cid.Hash) == 0
}

type Committer interface {

	// Commit persists the state to disk.
	Commit() CommitID
}

type CommitterLoader func(id CommitID) (Committer, error)

// A Store is anything that can be wrapped with a cache.
type CacheWrappable interface {

	// CacheWrap() wraps a thing with a cache.  After calling
	// .Write() on the CacheWrap, all previous CacheWraps on the
	// object expire.
	//
	// CacheWrap() should not return a Committer, since Commit() on
	// CacheWraps make no sense.  It can return KVStore, IterKVStore,
	// etc.
	//
	// NOTE: https://dave.cheney.net/2017/07/22/should-go-2-0-support-generics.
	// The returned object may or may not implement CacheWrap() as well.
	CacheWrap() interface{}
}

// KVStore is a simple interface to get/set data
type KVStore interface {
	CacheWrappable // CacheWrap returns KVStore

	Set(key, value []byte) (prev []byte)
	Get(key []byte) (value []byte, exists bool)
	Has(key []byte) (exists bool)
	Remove(key []byte) (prev []byte, removed bool)
}

// IterKVStore can be iterated on
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
type IterKVStore interface {
	KVStore // CacheWrap returns IterKVMap

	Iterator(start, end []byte) Iterator
	ReverseIterator(start, end []byte) Iterator

	First(start, end []byte) (kv KVPair, ok bool)
	Last(start, end []byte) (kv KVPair, ok bool)
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
