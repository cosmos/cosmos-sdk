package branch

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"
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

// set adds a new key-value pair to the change set's tree.
func (bt changeSet) set(key, value []byte) {
	bt.tree.Set(newItem(key, value))
}

// get retrieves the value associated with the given key from the changeSet's tree.
func (bt changeSet) get(key []byte) (value []byte, found bool) {
	it, found := bt.tree.Get(item{key: key})
	return it.value, found
}

// delete removes the value associated with the given key from the change set.
// If the key does not exist in the change set, this method does nothing.
func (bt changeSet) delete(key []byte) {
	bt.set(key, nil)
}

// iterator returns a new iterator over the key-value pairs in the changeSet
// that have keys greater than or equal to the start key and less than the end key.
func (bt changeSet) iterator(start, end []byte) (*memIterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	return newMemIterator(start, end, bt.tree, true), nil
}

// reverseIterator returns a new iterator that iterates over the key-value pairs in reverse order
// within the specified range [start, end) in the changeSet's tree.
// If start or end is an empty byte slice, it returns an error indicating that the key is empty.
func (bt changeSet) reverseIterator(start, end []byte) (*memIterator, error) {
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

// newMemIterator creates a new memory iterator for a given range of keys in a B-tree.
// The iterator creates a copy then starts at the specified start key and ends at the specified end key.
// The `tree` parameter is the B-tree to iterate over.
// The `ascending` parameter determines the direction of iteration.
// If `ascending` is true, the iterator will iterate in ascending order.
// If `ascending` is false, the iterator will iterate in descending order.
// The returned iterator is positioned at the first key that is greater than or equal to the start key.
// If the start key is nil, the iterator is positioned at the first key in the B-tree.
// If the end key is nil, the iterator is positioned at the last key in the B-tree.
// The iterator is inclusive of the start key and exclusive of the end key.
// The `valid` field of the iterator indicates whether the iterator is positioned at a valid key.
// The `start` and `end` fields of the iterator store the start and end keys respectively.
func newMemIterator(start, end []byte, tree *btree.BTreeG[item], ascending bool) *memIterator {
	iter := tree.Copy().Iter()
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

// Domain returns the start and end keys of the iterator's domain.
func (mi *memIterator) Domain() (start, end []byte) {
	return mi.start, mi.end
}

// Close releases any resources held by the iterator.
func (mi *memIterator) Close() error {
	mi.iter.Release()
	mi.valid = false
	return nil
}

// Error returns the error state of the iterator.
// If the iterator is not valid, it returns the errInvalidIterator error.
// Otherwise, it returns nil.
func (mi *memIterator) Error() error {
	if !mi.Valid() {
		return errInvalidIterator
	}
	return nil
}

// Valid returns whether the iterator is currently pointing to a valid entry.
// It returns true if the iterator is valid, and false otherwise.
func (mi *memIterator) Valid() bool {
	return mi.valid
}

// Next advances the iterator to the next key-value pair.
// If the iterator is in ascending order, it moves to the next key-value pair.
// If the iterator is in descending order, it moves to the previous key-value pair.
// It also checks if the new key-value pair is within the specified range.
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

// keyInRange checks if the given key is within the range defined by the iterator.
// If the iterator is in ascending order and the end key is not nil, it returns false
// if the key is greater than or equal to the end key.
// If the iterator is in descending order and the start key is not nil, it returns false
// if the key is less than the start key.
// Otherwise, it returns true.
func (mi *memIterator) keyInRange(key []byte) bool {
	if mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	if !mi.ascending && mi.start != nil && bytes.Compare(key, mi.start) < 0 {
		return false
	}
	if !mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	return true
}

// Key returns the key of the current item in the iterator.
func (mi *memIterator) Key() []byte {
	return mi.iter.Item().key
}

// Value returns the value of the current item in the iterator.
func (mi *memIterator) Value() []byte {
	return mi.iter.Item().value
}

// assertValid checks if the memIterator is in a valid state.
// If there is an error, it panics with the error message.
func (mi *memIterator) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}
