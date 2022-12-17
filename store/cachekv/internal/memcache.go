package internal

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"
)

var errKeyEmpty = errors.New("key cannot be empty")

const (
	// The approximate number of items and children per B-tree node. Tuned with benchmarks.
	// copied from memdb.
	bTreeDegree = 32
)

// MemCache implements the in-memory cache for cachekv store,
// we don't use MemDB here because cachekv is used extensively in sdk core path,
// we need it to be as fast as possible, while `MemDB` is mainly used as a mocking db in unit tests.
//
// We choose tidwall/btree over google/btree here because it provides API to implement step iterator directly.
type MemCache struct {
	tree *btree.BTreeG[item]
}

// NewMemCache creates a wrapper around `btree.BTreeG`.
func NewMemCache() MemCache {
	return MemCache{tree: btree.NewBTreeGOptions(byKeys, btree.Options{
		Degree:  bTreeDegree,
		NoLocks: false,
	})}
}

// Set set a cache entry, dirty means it's newly set,
// `nil` value means a deletion.
func (bt MemCache) Set(key, value []byte, dirty bool) {
	bt.tree.Set(item{key: key, value: value, dirty: dirty})
}

// Get returns (value, found)
func (bt MemCache) Get(key []byte) ([]byte, bool) {
	i, found := bt.tree.Get(item{key: key})
	if !found {
		return nil, false
	}
	return i.value, true
}

// ScanDirtyItems iterate over the dirty entries.
func (bt MemCache) ScanDirtyItems(fn func(key, value []byte)) {
	bt.tree.Scan(func(item item) bool {
		if item.dirty {
			fn(item.key, item.value)
		}
		return true
	})
}

// Copy the cache. This is a copy-on-write operation and is very fast because
// it only performs a shadowed copy.
func (bt MemCache) Copy() MemCache {
	return MemCache{tree: bt.tree.IsoCopy()}
}

// Iterator iterates on a isolated view of the cache, not affected by future modifications.
func (bt MemCache) Iterator(start, end []byte) *memIterator {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		panic(errKeyEmpty)
	}
	return NewMemIterator(start, end, bt.Copy(), true)
}

// ReverseIterator iterates on a isolated view of the cache, not affected by future modifications.
func (bt MemCache) ReverseIterator(start, end []byte) *memIterator {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		panic(errKeyEmpty)
	}
	return NewMemIterator(start, end, bt.Copy(), false)
}

// item represents a cached key-value pair and the entry of the cache btree.
// If dirty is true, it indicates the cached value is newly set, maybe different from the underlying value.
type item struct {
	key   []byte
	value []byte
	dirty bool
}

// byKeys compares the items by key
func byKeys(a, b item) bool {
	return bytes.Compare(a.key, b.key) == -1
}
