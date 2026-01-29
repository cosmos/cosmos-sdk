package cachekv

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
type BTree struct {
	tree *btree.BTreeG[item]
}

// NewBTree creates a wrapper around `btree.BTreeG`.
func NewBTree() BTree {
	return BTree{
		tree: btree.NewBTreeGOptions(byKeys, btree.Options{
			Degree:  bTreeDegree,
			NoLocks: false,
		}),
	}
}

func (bt BTree) Set(key, value []byte) {
	bt.tree.Set(newItem(key, value, true))
}

func (bt BTree) SetCached(key, value []byte) {
	bt.tree.Set(newItem(key, value, false))
}

func (bt BTree) Get(key []byte) ([]byte, bool) {
	i, found := bt.tree.Get(newItem(key, nil, false))
	if !found {
		return nil, false
	}
	return i.value, true
}

func (bt BTree) Delete(key []byte) {
	bt.tree.Set(newItem(key, nil, true))
}

func (bt BTree) Iterator(start, end []byte) (types.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt, true), nil
}

func (bt BTree) ReverseIterator(start, end []byte) (types.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt, false), nil
}

func (bt BTree) Scan(f func(key []byte, value []byte, dirty bool) bool) {
	bt.tree.Scan(func(item item) bool {
		return f(item.key, item.value, item.dirty)
	})
}

// Copy the tree. This is a copy-on-write operation and is very fast because
// it only performs a shadowed copy.
func (bt BTree) Copy() BTree {
	return BTree{
		tree: bt.tree.Copy(),
	}
}

func (bt BTree) Clear() {
	bt.tree.Clear()
}

// item is a btree item with byte slices as keys and values
type item struct {
	key   []byte
	value []byte
	dirty bool
}

// byKeys compares the items by key
func byKeys(a, b item) bool {
	return bytes.Compare(a.key, b.key) == -1
}

// newItem creates a new pair item.
func newItem(key, value []byte, dirty bool) item {
	return item{key: key, value: value, dirty: dirty}
}
