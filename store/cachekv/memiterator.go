package cachekv

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	types.Iterator

	deleted map[string]struct{}
}

func newMemIterator(start, end []byte, items *dbm.MemDB, deleted map[string]struct{}, ascending bool) *memIterator {
	var iter types.Iterator
	var err error

	if ascending {
		iter, err = items.Iterator(start, end)
	} else {
		iter, err = items.ReverseIterator(start, end)
	}

	if err != nil {
		panic(err)
	}

	newDeleted := make(map[string]struct{})
	for k, v := range deleted {
		newDeleted[k] = v
	}

	return &memIterator{
		Iterator: iter,

		deleted: newDeleted,
	}
}

func (mi *memIterator) Value() []byte {
	key := mi.Iterator.Key()
	if _, ok := mi.deleted[string(key)]; ok {
		return nil
	}
	return mi.Iterator.Value()
}
