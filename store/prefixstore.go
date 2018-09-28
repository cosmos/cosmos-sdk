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

func cloneAppend(bz []byte, bz2 []byte) (res []byte) {
	res = make([]byte, len(bz)+len(bz2))
	copy(res, bz)
	copy(res[len(bz):], bz2)
	return
}

func (s prefixStore) key(key []byte) (res []byte) {
	res = cloneAppend(s.prefix, key)
	return
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
	res := s.parent.Get(s.key(key))
	return res
}

// Implements KVStore
func (s prefixStore) Has(key []byte) bool {
	return s.parent.Has(s.key(key))
}

// Implements KVStore
func (s prefixStore) Set(key, value []byte) {
	s.parent.Set(s.key(key), value)
}

// Implements KVStore
func (s prefixStore) Delete(key []byte) {
	s.parent.Delete(s.key(key))
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
	newstart := cloneAppend(s.prefix, start)

	var newend []byte
	if end == nil {
		newend = sdk.PrefixEndBytes(s.prefix)
	} else {
		newend = cloneAppend(s.prefix, end)
	}

	return prefixIterator{
		prefix: s.prefix,
		iter:   s.parent.Iterator(newstart, newend),
	}
}

// Implements KVStore
func (s prefixStore) ReverseIterator(start, end []byte) Iterator {
	newstart := make([]byte, len(s.prefix), len(start))
	copy(newstart, s.prefix)
	newstart = append(newstart, start...)

	newend := make([]byte, len(s.prefix)+len(end))
	if end == nil {
		newend = sdk.PrefixEndBytes(s.prefix)
	} else {
		copy(newend, s.prefix)
		newend = append(newend, end...)
	}

	return prefixIterator{
		prefix: s.prefix,
		iter:   s.parent.ReverseIterator(newstart, newend),
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
