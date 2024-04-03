package btree

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"

	"cosmossdk.io/store/types"
)

const (
	// The approximate number of items and children per B-tree node. Tuned with benchmarks.
	// copied from memdb.
	bTreeDegree = 32
)

var errKeyEmpty = errors.New("key cannot be empty")

// BTree implements the sorted cache for cachekv store,
// we don't use MemDB here because cachekv is used extensively in sdk core path,
// we need it to be as fast as possible, while `MemDB` is mainly used as a mocking db in unit tests.
//
// We choose tidwall/btree over google/btree here because it provides API to implement step iterator directly.
type BTree[V any] struct {
	tree *btree.BTreeG[item[V]]
}

// NewBTree creates a wrapper around `btree.BTreeG`.
func NewBTree[V any]() BTree[V] {
	return BTree[V]{}
}

func (bt *BTree[V]) init() {
	if bt.tree != nil {
		return
	}
	bt.tree = btree.NewBTreeGOptions(byKeys[V], btree.Options{
		Degree:  bTreeDegree,
		NoLocks: false,
	})
}

// Set supports nil as value when used as overlay
func (bt *BTree[V]) Set(key []byte, value V) {
	bt.init()
	bt.tree.Set(newItem(key, value))
}

func (bt BTree[V]) Get(key []byte) (V, bool) {
	if bt.tree == nil {
		var zero V
		return zero, false
	}
	i, found := bt.tree.Get(newItemWithKey[V](key))
	return i.value, found
}

func (bt *BTree[V]) Delete(key []byte) {
	if bt.tree == nil {
		return
	}
	bt.tree.Delete(newItemWithKey[V](key))
}

func (bt BTree[V]) Iterator(start, end []byte) (types.GIterator[V], error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt, true), nil
}

func (bt BTree[V]) ReverseIterator(start, end []byte) (types.GIterator[V], error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt, false), nil
}

// Copy the tree. This is a copy-on-write operation and is very fast because
// it only performs a shadowed copy.
func (bt BTree[V]) Copy() BTree[V] {
	if bt.tree == nil {
		return BTree[V]{}
	}
	return BTree[V]{
		tree: bt.tree.Copy(),
	}
}

func (bt BTree[V]) Clear() {
	if bt.tree == nil {
		return
	}
	bt.tree.Clear()
}

func (bt BTree[V]) Scan(cb func(key []byte, value V) bool) {
	if bt.tree == nil {
		return
	}
	bt.tree.Scan(func(i item[V]) bool {
		return cb(i.key, i.value)
	})
}

// item is a btree item with byte slices as keys and values
type item[V any] struct {
	key   []byte
	value V
}

// byKeys compares the items by key
func byKeys[V any](a, b item[V]) bool {
	return bytes.Compare(a.key, b.key) == -1
}

// newItem creates a new pair item.
func newItem[V any](key []byte, value V) item[V] {
	return item[V]{key: key, value: value}
}

// newItem creates a new pair item with empty value.
func newItemWithKey[V any](key []byte) item[V] {
	return item[V]{key: key}
}
