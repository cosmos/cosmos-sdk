package branch

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"

	"cosmossdk.io/core/store"
)

const (
	// The approximate number of items and children per B-tree node. Tuned with benchmarks.
	// copied from memdb.
	bTreeDegree = 32
)

var errKeyEmpty = errors.New("key cannot be empty")

// changeSet implements the sorted cache for cachekv store,
// we don't use MemDB here because cachekv is used extensively in sdk core path,
// we need it to be as fast as possible, while `MemDB` is mainly used as a mocking db in unit tests.
//
// We choose tidwall/btree over google/btree here because it provides API to implement step iterator directly.
type changeSet struct {
	tree *btree.BTreeG[item]
}

// newChangeSet creates a wrapper around `btree.BTreeG`.
func newChangeSet() changeSet {
	return changeSet{
		tree: btree.NewBTreeGOptions(byKeys, btree.Options{
			Degree:  bTreeDegree,
			NoLocks: true,
		}),
	}
}

func (bt changeSet) set(key, value []byte) {
	bt.tree.Set(newItem(key, value))
}

func (bt changeSet) get(key []byte) (value []byte, found bool) {
	it, found := bt.tree.Get(item{key: key})
	return it.value, found
}

func (bt changeSet) delete(key []byte) {
	bt.set(key, nil)
}

func (bt changeSet) iterator(start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt.tree, true), nil
}

func (bt changeSet) reverseIterator(start, end []byte) (store.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt.tree, false), nil
}

// item is a btree item with byte slices as keys and values
type item struct {
	key   []byte
	value []byte
}

// byKeys compares the items by key
func byKeys(a, b item) bool {
	return bytes.Compare(a.key, b.key) == -1
}

// newItem creates a new pair item.
func newItem(key, value []byte) item {
	return item{key: key, value: value}
}

// memIterator iterates over iterKVCache items.
// if value is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	iter btree.IterG[item]

	start     []byte
	end       []byte
	ascending bool
	valid     bool
}

func newMemIterator(start, end []byte, tree *btree.BTreeG[item], ascending bool) *memIterator {
	iter := tree.Iter()
	var valid bool
	if ascending {
		if start != nil {
			valid = iter.Seek(newItem(start, nil))
		} else {
			valid = iter.First()
		}
	} else {
		if end != nil {
			valid = iter.Seek(newItem(end, nil))
			if !valid {
				valid = iter.Last()
			} else {
				// end is exclusive
				valid = iter.Prev()
			}
		} else {
			valid = iter.Last()
		}
	}

	mi := &memIterator{
		iter:      iter,
		start:     start,
		end:       end,
		ascending: ascending,
		valid:     valid,
	}

	if mi.valid {
		mi.valid = mi.keyInRange(mi.Key())
	}

	return mi
}

func (mi *memIterator) Domain() (start, end []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Close() error {
	mi.iter.Release()
	return nil
}

func (mi *memIterator) Error() error {
	if !mi.Valid() {
		return errInvalidIterator
	}
	return nil
}

func (mi *memIterator) Valid() bool {
	return mi.valid
}

func (mi *memIterator) Next() {
	mi.assertValid()

	if mi.ascending {
		mi.valid = mi.iter.Next()
	} else {
		mi.valid = mi.iter.Prev()
	}

	if mi.valid {
		mi.valid = mi.keyInRange(mi.Key())
	}
}

func (mi *memIterator) keyInRange(key []byte) bool {
	if mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	if !mi.ascending && mi.start != nil && bytes.Compare(key, mi.start) < 0 {
		return false
	}
	return true
}

func (mi *memIterator) Key() []byte {
	return mi.iter.Item().key
}

func (mi *memIterator) Value() []byte {
	return mi.iter.Item().value
}

func (mi *memIterator) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}
