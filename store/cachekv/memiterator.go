package cachekv

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"
)

// memIterator iterates over iterKVCache items.
// if value is nil, means it was deleted.
// Implements Iterator.
type memIterator[V any] struct {
	iter btree.MapIter[string, V]

	start     []byte
	end       []byte
	ascending bool
	valid     bool
}

func newMemIterator[V any](start, end []byte, items *btree.Map[string, V], ascending bool) *memIterator[V] {
	items = items.Copy() // copy the btree to avoid concurrent modification issues, this should be O(1) due to copy-on-write semantics of the btree
	iter := items.Iter()
	var valid bool
	if ascending {
		if start != nil {
			// we use unsafeBytesToString here because we are doing a lookup, we are not modifying the key, and we want to avoid unnecessary allocations
			valid = iter.Seek(unsafeBytesToString(start))
		} else {
			valid = iter.First()
		}
	} else {
		if end != nil {
			// we use unsafeBytesToString here because we are doing a lookup
			valid = iter.Seek(unsafeBytesToString(end))
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
	return []byte(mi.iter.Key()) // this introduces a small amount of allocation and copying, but is safer
}

func (mi *memIterator[V]) Value() V {
	return mi.iter.Value()
}

func (mi *memIterator[V]) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}
