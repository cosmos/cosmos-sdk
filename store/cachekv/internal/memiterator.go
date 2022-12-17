package internal

import (
	"bytes"
	"errors"

	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/tidwall/btree"
)

var _ types.Iterator = (*MemIterator)(nil)

// memIterator iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type MemIterator struct {
	iter btree.IterG[item]

	start     []byte
	end       []byte
	ascending bool
	lastKey   []byte
	deleted   map[string]struct{}
	valid     bool
}

func NewMemIterator(start, end []byte, items *BTree, deleted map[string]struct{}, ascending bool) *MemIterator {
	iter := items.tree.Iter()
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

	mi := &MemIterator{
		iter:      iter,
		start:     start,
		end:       end,
		ascending: ascending,
		lastKey:   nil,
		deleted:   deleted,
		valid:     valid,
	}

	if mi.valid {
		mi.valid = mi.keyInRange(mi.Key())
	}

	return mi
}

func (mi *MemIterator) Domain() (start []byte, end []byte) {
	return mi.start, mi.end
}

func (mi *MemIterator) Close() error {
	mi.iter.Release()
	return nil
}

func (mi *MemIterator) Error() error {
	if !mi.Valid() {
		return errors.New("invalid memIterator")
	}
	return nil
}

func (mi *MemIterator) Valid() bool {
	return mi.valid
}

func (mi *MemIterator) Next() {
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

func (mi *MemIterator) keyInRange(key []byte) bool {
	if mi.ascending && mi.end != nil && bytes.Compare(key, mi.end) >= 0 {
		return false
	}
	if !mi.ascending && mi.start != nil && bytes.Compare(key, mi.start) < 0 {
		return false
	}
	return true
}

func (mi *MemIterator) Key() []byte {
	return mi.iter.Item().key
}

func (mi *MemIterator) Value() []byte {
	item := mi.iter.Item()
	key := item.key
	// We need to handle the case where deleted is modified and includes our current key
	// We handle this by maintaining a lastKey object in the iterator.
	// If the current key is the same as the last key (and last key is not nil / the start)
	// then we are calling value on the same thing as last time.
	// Therefore we don't check the mi.deleted to see if this key is included in there.
	if _, ok := mi.deleted[string(key)]; ok {
		if mi.lastKey == nil || !bytes.Equal(key, mi.lastKey) {
			// not re-calling on old last key
			return nil
		}
	}
	mi.lastKey = key
	return item.value
}

func (mi *MemIterator) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}
