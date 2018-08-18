package store

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ KVStore = prefixStore{}

type prefixStore struct {
	parent KVStore
	prefix []byte
}

// Implements Store
func (s prefixStore) GetStoreType() StoreType {
	return s.parent.GetStoreType()
}

// Implements CacheWrap
func (s prefixStore) CacheWrap() CacheWrap {
	return NewCacheKVStore(s)
}

// CacheWrapWithTrace implements the KVStore interface.
func (s prefixStore) CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap {
	return NewCacheKVStore(NewTraceKVStore(s, w, tc))
}

// Implements KVStore
func (s prefixStore) Get(key []byte) []byte {
	return s.parent.Get(append(s.prefix, key...))
}

// Implements KVStore
func (s prefixStore) Has(key []byte) bool {
	return s.parent.Has(append(s.prefix, key...))
}

// Implements KVStore
func (s prefixStore) Set(key, value []byte) {
	s.parent.Set(append(s.prefix, key...), value)
}

// Implements KVStore
func (s prefixStore) Delete(key []byte) {
	s.parent.Delete(append(s.prefix, key...))
}

// Implements KVStore
func (s prefixStore) Prefix(prefix []byte) KVStore {
	return prefixStore{s, prefix}
}

// Implements KVStore
func (s prefixStore) Gas(meter GasMeter, config GasConfig) KVStore {
	return NewGasKVStore(meter, config, s)
}

// Implements KVStore
func (s prefixStore) Iterator(start, end []byte) Iterator {
	if end == nil {
		end = sdk.PrefixEndBytes(s.prefix)
	} else {
		end = append(s.prefix, end...)
	}
	return prefixIterator{
		prefix: s.prefix,
		iter:   s.parent.Iterator(append(s.prefix, start...), end),
	}
}

// Implements KVStore
func (s prefixStore) ReverseIterator(start, end []byte) Iterator {
	if end == nil {
		end = sdk.PrefixEndBytes(s.prefix)
	} else {
		end = append(s.prefix, end...)
	}
	return prefixIterator{
		prefix: s.prefix,
		iter:   s.parent.ReverseIterator(start, end),
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
