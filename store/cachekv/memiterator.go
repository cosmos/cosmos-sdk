package cachekv

import (
	"container/list"
	"errors"

	tmkv "github.com/tendermint/tendermint/libs/kv"
	dbm "github.com/tendermint/tm-db"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	start, end []byte
	items      []*tmkv.Pair
	ascending  bool
}

func newMemIterator(start, end []byte, items *list.List, ascending bool) *memIterator {
	itemsInDomain := make([]*tmkv.Pair, 0)
	var entered bool
	for e := items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*tmkv.Pair)
		if !dbm.IsKeyInDomain(item.Key, start, end) {
			if entered {
				break
			}
			continue
		}
		itemsInDomain = append(itemsInDomain, item)
		entered = true
	}

	return &memIterator{
		start:     start,
		end:       end,
		items:     itemsInDomain,
		ascending: ascending,
	}
}

func (mi *memIterator) Domain() ([]byte, []byte) {
	return mi.start, mi.end
}

func (mi *memIterator) Valid() bool {
	return len(mi.items) > 0
}

func (mi *memIterator) assertValid() {
	if err := mi.Error(); err != nil {
		panic(err)
	}
}

func (mi *memIterator) Next() {
	mi.assertValid()
	if mi.ascending {
		mi.items = mi.items[1:]
	} else {
		mi.items = mi.items[:len(mi.items)-1]
	}
}

func (mi *memIterator) Key() []byte {
	mi.assertValid()
	if mi.ascending {
		return mi.items[0].Key
	}
	return mi.items[len(mi.items)-1].Key
}

func (mi *memIterator) Value() []byte {
	mi.assertValid()
	if mi.ascending {
		return mi.items[0].Value
	}
	return mi.items[len(mi.items)-1].Value
}

func (mi *memIterator) Close() {
	mi.start = nil
	mi.end = nil
	mi.items = nil
}

// Error returns an error if the memIterator is invalid defined by the Valid
// method.
func (mi *memIterator) Error() error {
	if !mi.Valid() {
		return errors.New("invalid memIterator")
	}

	return nil
}
