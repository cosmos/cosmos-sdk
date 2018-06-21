package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type prefixStore struct {
	store  KVStore
	prefix []byte
}

// Implements Store
func (s prefixStore) GetStoreType() StoreType {
	return sdk.StoreTypePrefix
}

// Implements CacheWrap
func (s prefixStore) CacheWrap() CacheWrap {
	return NewCacheKVStore(s)
}

// Implements KVStore
func (s prefixStore) Get(key []byte) []byte {
	return s.store.Get(append(s.prefix, key...))
}

// Implements KVStore
func (s prefixStore) Has(key []byte) bool {
	return s.store.Has(append(s.prefix, key...))
}

// Implements KVStore
func (s prefixStore) Set(key, value []byte) {
	s.store.Set(append(s.prefix, key...), value)
}

// Implements KVStore
func (s prefixStore) Delete(key []byte) {
	s.store.Delete(append(s.prefix, key...))
}

// Implements KVStore
func (s prefixStore) Prefix(prefix []byte) KVStore {
	return prefixStore{s, prefix}
}

// Implements KVStore
func (s prefixStore) Iterator(start, end []byte) Iterator {
	return prefixIterator{
		prefix: s.prefix,
		iter:   s.store.Iterator(start, end),
	}
}

// Implements KVStore
func (s prefixStore) ReverseIterator(start, end []byte) Iterator {
	return prefixIterator{
		prefix: s.prefix,
		iter:   s.store.ReverseIterator(start, end),
	}
}

type prefixIterator struct {
	prefix []byte

	iter Iterator
}

// Implements Iterator
func (iter prefixIterator) Domain() (start []byte, end []byte) {
	start, end = iter.iter.Domain()
	start = start[len(iter.prefix):]
	end = end[len(iter.prefix):]
	return
}

// Implements Iterator
func (iter prefixIterator) Valid() bool {
	return iter.iter.Valid()
}

// Implements Iterator
func (iter prefixIterator) Next() {
	iter.iter.Next()
}

// Implements Iterator
func (iter prefixIterator) Key() (key []byte) {
	key = iter.iter.Key()
	key = key[len(iter.prefix):]
	return
}

// Implements Iterator
func (iter prefixIterator) Value() []byte {
	return iter.iter.Value()
}

// Implements Iterator
func (iter prefixIterator) Close() {
	iter.iter.Close()
}
