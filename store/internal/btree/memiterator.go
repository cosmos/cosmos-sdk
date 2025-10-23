package btree

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"

	"cosmossdk.io/store/types"
)

var _ types.Iterator = (*memIterator[[]byte])(nil)

// memIterator iterates over iterKVCache items.
// if value is nil, means it was deleted.
// Implements Iterator.
type memIterator[V any] struct {
	iter btree.IterG[item[V]]

	start     []byte
	end       []byte
	ascending bool
	valid     bool
}

func newMemIterator[V any](start, end []byte, items BTree[V], ascending bool) *memIterator[V] {
	var (
		valid bool
		empty V
	)
	iter := items.tree.Iter()
	if ascending {
		if start != nil {
			valid = iter.Seek(newItem(start, empty))
		} else {
			valid = iter.First()
		}
	} else {
		if end != nil {
			valid = iter.Seek(newItem(end, empty))
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

	mi := &memIterator[V]{
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

func (mi *memIterator[V]) Domain() (start, end []byte) {
	return mi.start, mi.end
}

func (mi *memIterator[V]) Close() error {
	mi.iter.Release()
	return nil
}

func (mi *memIterator[V]) Error() error {
	if !mi.Valid() {
		return errors.New("invalid memIterator")
	}
	return nil
}

func (mi *memIterator[V]) Valid() bool {
	return mi.valid
}

func (mi *memIterator[V]) Next() {
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

func (mi *memIterator[V]) keyInRange(key []byte) bool {
	if mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	if !mi.ascending && mi.start != nil && bytes.Compare(key, mi.start) < 0 {
		return false
	}
	return true
}

func (mi *memIterator[V]) Key() []byte {
	return mi.iter.Item().key
}

func (mi *memIterator[V]) Value() V {
	return mi.iter.Item().value
}

func (mi *memIterator[V]) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}
