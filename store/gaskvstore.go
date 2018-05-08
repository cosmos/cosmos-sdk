package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// gasKVStore applies gas tracking to an underlying kvstore
type gasKVStore struct {
	gasMeter sdk.GasMeter
	parent   KVStore
}

// nolint
func NewGasKVStore(gasMeter sdk.GasMeter, parent KVStore) *gasKVStore {
	kvs := &gasKVStore{
		gasMeter: gasMeter,
		parent:   parent,
	}
	return kvs
}

// Implements Store.
func (gi *gasKVStore) GetStoreType() StoreType {
	return gi.parent.GetStoreType()
}

// Implements KVStore.
func (gi *gasKVStore) Get(key []byte) (value []byte) {
	return gi.parent.Get(key)
}

// Implements KVStore.
func (gi *gasKVStore) Set(key []byte, value []byte) {
	gi.parent.Set(key, value)
}

// Implements KVStore.
func (gi *gasKVStore) Has(key []byte) bool {
	return gi.parent.Has(key)
}

// Implements KVStore.
func (gi *gasKVStore) Delete(key []byte) {
	gi.parent.Delete(key)
}

// Implements KVStore.
func (gi *gasKVStore) Iterator(start, end []byte) Iterator {
	return gi.iterator(start, end, true)
}

// Implements KVStore.
func (gi *gasKVStore) ReverseIterator(start, end []byte) Iterator {
	return gi.iterator(start, end, false)
}

// Implements KVStore.
func (gi *gasKVStore) SubspaceIterator(prefix []byte) Iterator {
	return gi.iterator(prefix, sdk.PrefixEndBytes(prefix), true)
}

// Implements KVStore.
func (gi *gasKVStore) ReverseSubspaceIterator(prefix []byte) Iterator {
	return gi.iterator(prefix, sdk.PrefixEndBytes(prefix), false)
}

// Implements KVStore.
func (gi *gasKVStore) CacheWrap() CacheWrap {
	return gi.parent.CacheWrap() // TODO
}

func (gi *gasKVStore) iterator(start, end []byte, ascending bool) Iterator {
	var parent Iterator
	if ascending {
		parent = gi.parent.Iterator(start, end)
	} else {
		parent = gi.parent.ReverseIterator(start, end)
	}
	return parent // TODO
}
