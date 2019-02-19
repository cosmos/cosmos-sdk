package cachekv

import (
	"bytes"

	dbm "github.com/tendermint/tendermint/libs/db"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	start, end []byte
	items      *heap
}

func newMemIterator(start, end []byte, items *heap) *memIterator {
	return &memIterator{
		start: start,
		end:   end,
		items: items,
	}
}

func (mi *memIterator) Domain() ([]byte, []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Valid() bool {
	if mi.items == nil {
		return false
	}

	return !mi.items.isEmpty()
}

func (mi *memIterator) assertValid() {
	if !mi.Valid() {
		panic("memIterator is invalid")
	}
}

func (mi *memIterator) Next() {
	mi.assertValid()
	mi.items.pop()
	if mi.Valid() {
		if !dbm.IsKeyInDomain(mi.Key(), mi.start, mi.end) {
			mi.items = nil
		}
	}
}

func (mi *memIterator) Key() []byte {
	mi.assertValid()
	return mi.items.peek().Key
}

func (mi *memIterator) Value() []byte {
	mi.assertValid()
	return mi.items.peek().Value
}

func (mi *memIterator) Close() {
	mi.start = nil
	mi.end = nil
	mi.items = nil
}

//----------------------------------------
// Misc.

// bytes.Compare but bounded on both sides by nil.
// both (k1, nil) and (nil, k2) return -1
func keyCompare(k1, k2 []byte) int {
	if k1 == nil && k2 == nil {
		return 0
	} else if k1 == nil || k2 == nil {
		return -1
	}
	return bytes.Compare(k1, k2)
}
