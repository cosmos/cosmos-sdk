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
	return BTree[V]{
		tree: btree.NewBTreeGOptions(byKeys[V], btree.Options{
			Degree:  bTreeDegree,
			NoLocks: false,
		}),
	}
}

func (bt BTree[V]) Set(key []byte, value V) {
	bt.tree.Set(newItem(key, value))
}

func (bt BTree[V]) Get(key []byte) V {
	var empty V
	i, found := bt.tree.Get(newItem(key, empty))
	if !found {
		return empty
	}
	return i.value
}

func (bt BTree[V]) Delete(key []byte) {
	var empty V
	bt.tree.Delete(newItem(key, empty))
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
	return BTree[V]{
		tree: bt.tree.Copy(),
	}
}

func (bt BTree[V]) Clear() {
	bt.tree.Clear()
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
