package store

import dbm "github.com/tendermint/tmlibs/db"

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	start, end []byte
	items      []KVPair
}

func newMemIterator(start, end []byte, items []KVPair) *memIterator {
	itemsInDomain := make([]KVPair, 0)
	for _, item := range items {
		if dbm.IsKeyInDomain(string(item.Key), start, end) {
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

func (mi *memIterator) Release() {
	mi.start = nil
	mi.end = nil
	mi.items = nil
}
