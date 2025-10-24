package tree

import (
	"bytes"
	"errors"

	"github.com/tidwall/btree"
)

// BTreeIteratorG iterates over btree.
// Implements Iterator.
type BTreeIteratorG[T KeyItem] struct {
	iter btree.IterG[T]

	start     []byte
	end       []byte
	ascending bool
	valid     bool
}

func NewNoopBTreeIteratorG[T KeyItem](
	start, end []byte,
	ascending bool,
	valid bool,
) BTreeIteratorG[T] {
	return BTreeIteratorG[T]{
		start:     start,
		end:       end,
		ascending: ascending,
		valid:     valid,
	}
}

func NewBTreeIteratorG[T KeyItem](
	startItem, endItem T,
	iter btree.IterG[T],
	ascending bool,
) *BTreeIteratorG[T] {
	start := startItem.GetKey()
	end := endItem.GetKey()

	var valid bool
	if ascending {
		if start != nil {
			valid = iter.Seek(startItem)
		} else {
			valid = iter.First()
		}
	} else {
		if end != nil {
			valid = iter.Seek(endItem)
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

	mi := &BTreeIteratorG[T]{
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

func (mi *BTreeIteratorG[T]) Domain() (start, end []byte) {
	return mi.start, mi.end
}

func (mi *BTreeIteratorG[T]) Close() error {
	mi.iter.Release()
	return nil
}

func (mi *BTreeIteratorG[T]) Error() error {
	if !mi.Valid() {
		return errors.New("invalid memIterator")
	}
	return nil
}

func (mi *BTreeIteratorG[T]) Valid() bool {
	return mi.valid
}

func (mi *BTreeIteratorG[T]) Invalidate() {
	mi.valid = false
}

func (mi *BTreeIteratorG[T]) Next() {
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

func (mi *BTreeIteratorG[T]) keyInRange(key []byte) bool {
	if mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	if !mi.ascending && mi.start != nil && bytes.Compare(key, mi.start) < 0 {
		return false
	}
	return true
}

func (mi *BTreeIteratorG[T]) Item() T {
	return mi.iter.Item()
}

func (mi *BTreeIteratorG[T]) Key() []byte {
	return mi.Item().GetKey()
}

func (mi *BTreeIteratorG[T]) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}
