package internal

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/tidwall/btree"
)

var _ types.Iterator = &memIterator{}

// memIterator iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	iter btree.GenericIter[item]

	start     []byte
	end       []byte
	ascending bool
	lastKey   []byte
	deleted   map[string]struct{}
	valid     bool
}

func NewMemIterator(start, end []byte, items *BTree, deleted map[string]struct{}, ascending bool) *memIterator {
	iter := items.tree.Iter()
	var valid bool
	if ascending {
		if start != nil {
			valid = iter.Seek(newKey(start))
		} else {
			valid = iter.First()
		}
	} else {
		if end != nil {
			valid = iter.Seek(newKey(end))
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
	return &memIterator{
		iter:      iter,
		start:     start,
		end:       end,
		ascending: ascending,
		lastKey:   nil,
		deleted:   deleted,
		valid:     valid,
	}
}

func (mi *memIterator) Domain() (start []byte, end []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Close() error {
	mi.iter.Release()
	return nil
}

func (mi *memIterator) Error() error {
	return nil
}

func (mi *memIterator) Valid() bool {
	if !mi.valid {
		return false
	}
	key := mi.iter.Item().key
	if mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	if !mi.ascending && mi.start != nil && bytes.Compare(key, mi.start) < 0 {
		return false
	}
	return true
}

func (mi *memIterator) Next() {
	if mi.ascending {
		mi.valid = mi.iter.Next()
	} else {
		mi.valid = mi.iter.Prev()
	}
}

func (mi *memIterator) Key() []byte {
	return mi.iter.Item().key
}

func (mi *memIterator) Value() []byte {
	item := mi.iter.Item()
	key := item.key
	// We need to handle the case where deleted is modified and includes our current key
	// We handle this by maintaining a lastKey object in the iterator.
	// If the current key is the same as the last key (and last key is not nil / the start)
	// then we are calling value on the same thing as last time.
	// Therefore we don't check the mi.deleted to see if this key is included in there.
	reCallingOnOldLastKey := (mi.lastKey != nil) && bytes.Equal(key, mi.lastKey)
	if _, ok := mi.deleted[string(key)]; ok && !reCallingOnOldLastKey {
		return nil
	}
	mi.lastKey = key
	return item.value
}
