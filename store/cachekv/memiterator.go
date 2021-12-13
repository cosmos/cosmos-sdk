package cachekv

import (
	"bytes"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// Iterates over iterKVCache items.
// if key is nil, means it was deleted.
// Implements Iterator.
type memIterator struct {
	types.Iterator

	lastKey []byte
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

	return &memIterator{
		Iterator: iter,

		lastKey: nil,
		deleted: deleted,
	}
}

func (mi *memIterator) Value() []byte {
	key := mi.Iterator.Key()
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
	return mi.Iterator.Value()
}
