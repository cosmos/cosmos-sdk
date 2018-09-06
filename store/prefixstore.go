package store

import (
	"bytes"
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ KVStore = prefixStore{}

type prefixStore struct {
	parent KVStore
	prefix []byte
}

func clone(bz []byte) (res []byte) {
	res = make([]byte, len(bz))
	copy(res, bz)
	return
}

func (s prefixStore) key(key []byte) (res []byte) {
	res = clone(s.prefix)
	res = append(res, key...)
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
// Check https://github.com/tendermint/tendermint/blob/master/libs/db/prefix_db.go#L106
func (s prefixStore) Iterator(start, end []byte) Iterator {
	newstart := append(clone(s.prefix), start...)

	var newend []byte
	if end == nil {
		newend = cpIncr(s.prefix)
	} else {
		newend = append(clone(s.prefix), end...)
	}

	return prefixIterator{
		prefix: s.prefix,
		start:  newstart,
		end:    newend,
		iter:   s.parent.Iterator(newstart, newend),
	}

}

// Implements KVStore
// Check https://github.com/tendermint/tendermint/blob/master/libs/db/prefix_db.go#L129
func (s prefixStore) ReverseIterator(start, end []byte) Iterator {
	var newstart []byte
	if start == nil {
		newstart = cpIncr(s.prefix)
	} else {
		newstart = append(clone(s.prefix), start...)
	}

	var newend []byte
	if end == nil {
		newend = cpIncr(s.prefix)
	} else {
		newend = append(clone(s.prefix), end...)
	}

	iter := s.parent.ReverseIterator(newstart, newend)
	if start == nil {
		skipOne(iter, cpIncr(s.prefix))
	}

	return prefixIterator{
		prefix: s.prefix,
		start:  newstart,
		end:    newend,
		iter:   iter,
	}
}

type prefixIterator struct {
	prefix     []byte
	start, end []byte
	iter       Iterator
	valid      bool
}

// Implements Iterator
func (iter prefixIterator) Domain() ([]byte, []byte) {
	return iter.start, iter.end
}

// Implements Iterator
func (iter prefixIterator) Valid() bool {
	return iter.valid && iter.iter.Valid()
}

// Implements Iterator
func (iter prefixIterator) Next() {
	if !iter.valid {
		panic("prefixIterator invalid, cannot call Next()")
	}
	iter.iter.Next()
	if !iter.iter.Valid() || !bytes.HasPrefix(iter.iter.Key(), iter.prefix) {
		iter.iter.Close()
		iter.valid = false
	}
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

// copied from github.com/tendermint/tendermint/libs/db/prefix_db.go
func stripPrefix(key []byte, prefix []byte) []byte {
	if len(key) < len(prefix) || !bytes.Equal(key[:len(prefix)], prefix) {
		panic("should not happen")
	}
	return key[len(prefix):]
}

// wrapping sdk.PrefixEndBytes
func cpIncr(bz []byte) []byte {
	return sdk.PrefixEndBytes(bz)
}

// copied from github.com/tendermint/tendermint/libs/db/util.go
func cpDecr(bz []byte) (ret []byte) {
	if len(bz) == 0 {
		panic("cpDecr expects non-zero bz length")
	}
	ret = clone(bz)
	for i := len(bz) - 1; i >= 0; i-- {
		if ret[i] > byte(0x00) {
			ret[i]--
			return
		}
		ret[i] = byte(0xFF)
		if i == 0 {
			return nil
		}
	}
	return nil
}

func skipOne(iter Iterator, skipKey []byte) {
	if iter.Valid() {
		if bytes.Equal(iter.Key(), skipKey) {
			iter.Next()
		}
	}
}
