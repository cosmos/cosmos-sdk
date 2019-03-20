package cachekv

import (
	"bytes"

	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	start, end []byte
	items      []cmn.KVPair
}

func newMemIterator(start, end []byte, items []cmn.KVPair) *memIterator {
	itemsInDomain := make([]cmn.KVPair, 0)
	for _, item := range items {
		if dbm.IsKeyInDomain(item.Key, start, end) {
			itemsInDomain = append(itemsInDomain, item)
		}
	}
	return &memIterator{
		start: start,
		end:   end,
		items: itemsInDomain,
	}
}

func (mi *memIterator) Domain() ([]byte, []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Valid() bool {
	return len(mi.items) > 0
}

func (mi *memIterator) assertValid() {
	if !mi.Valid() {
		panic("memIterator is invalid")
	}
}

func (mi *memIterator) Next() {
	mi.assertValid()
	mi.items = mi.items[1:]
}

func (mi *memIterator) Key() []byte {
	mi.assertValid()
	return mi.items[0].Key
}

func (mi *memIterator) Value() []byte {
	mi.assertValid()
	return mi.items[0].Value
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
