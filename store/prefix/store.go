package prefix

import (
	"bytes"
	"errors"
	"io"

	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/tracekv"
	"cosmossdk.io/store/types"
)

type (
	Store    = GStore[[]byte]
	ObjStore = GStore[any]
)

var (
	_ types.KVStore    = Store{}
	_ types.ObjKVStore = ObjStore{}
)

func NewStore(parent types.KVStore, prefix []byte) Store {
	return NewGStore(
		parent, prefix,
		types.BytesIsZero,
		types.BytesValueLen,
	)
}

func NewObjStore(parent types.ObjKVStore, prefix []byte) ObjStore {
	return NewGStore(
		parent, prefix,
		types.AnyIsZero,
		types.AnyValueLen,
	)
}

// GStore is similar to cometbft/cometbft/libs/db/prefix_db
// both give access only to the limited subset of the store
// for convenience or safety
type GStore[V any] struct {
	parent types.GKVStore[V]
	prefix []byte

	isZero   func(V) bool
	valueLen func(V) int
}

func NewGStore[V any](
	parent types.GKVStore[V], prefix []byte,
	isZero func(V) bool, valueLen func(V) int,
) GStore[V] {
	return GStore[V]{
		parent: parent,
		prefix: prefix,

		isZero:   isZero,
		valueLen: valueLen,
	}
}

func cloneAppend(bz, tail []byte) (res []byte) {
	res = make([]byte, len(bz)+len(tail))
	copy(res, bz)
	copy(res[len(bz):], tail)
	return res
}

func (s GStore[V]) key(key []byte) (res []byte) {
	if key == nil {
		panic("nil key on Store")
	}
	res = cloneAppend(s.prefix, key)
	return res
}

// GetStoreType implements Store, returning the parent store's type
func (s GStore[V]) GetStoreType() types.StoreType {
	return s.parent.GetStoreType()
}

// CacheWrap implements CacheWrap, returning a new CacheWrap with the parent store as the underlying store
func (s GStore[V]) CacheWrap() types.CacheWrap {
	return cachekv.NewGStore(s, s.isZero, s.valueLen)
}

// CacheWrapWithTrace implements the KVStore interface.
func (s GStore[V]) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	// We need to make a type assertion here as the tracekv store requires bytes value types for serialization.
	if store, ok := any(s).(*GStore[[]byte]); ok {
		return cachekv.NewGStore(tracekv.NewStore(store, w, tc), store.isZero, store.valueLen)
	}
	return s.CacheWrap()
}

// Get implements KVStore, calls Get on the parent store with the key prefixed with the prefix
func (s GStore[V]) Get(key []byte) V {
	res := s.parent.Get(s.key(key))
	return res
}

// Has implements KVStore, calls Has on the parent store with the key prefixed with the prefix
func (s GStore[V]) Has(key []byte) bool {
	return s.parent.Has(s.key(key))
}

// Set implements KVStore, calls Set on the parent store with the key prefixed with the prefix
func (s GStore[V]) Set(key []byte, value V) {
	types.AssertValidKey(key)
	types.AssertValidValueGeneric(value, s.isZero, s.valueLen)
	s.parent.Set(s.key(key), value)
}

// Delete implements KVStore, calls Delete on the parent store with the key prefixed with the prefix
func (s GStore[V]) Delete(key []byte) {
	s.parent.Delete(s.key(key))
}

// Iterator implements KVStore
// Check https://github.com/cometbft/cometbft-db/blob/main/prefixdb_iterator.go#L106
func (s GStore[V]) Iterator(start, end []byte) types.GIterator[V] {
	newStart := cloneAppend(s.prefix, start)

	var newEnd []byte
	if end == nil {
		newEnd = cpIncr(s.prefix)
	} else {
		newEnd = cloneAppend(s.prefix, end)
	}

	iter := s.parent.Iterator(newStart, newEnd)

	return newPrefixIterator(s.prefix, start, end, iter)
}

// ReverseIterator implements KVStore
// Check https://github.com/cometbft/cometbft-db/blob/main/prefixdb_iterator.go#L129
func (s GStore[V]) ReverseIterator(start, end []byte) types.GIterator[V] {
	newstart := cloneAppend(s.prefix, start)

	var newend []byte
	if end == nil {
		newend = cpIncr(s.prefix)
	} else {
		newend = cloneAppend(s.prefix, end)
	}

	iter := s.parent.ReverseIterator(newstart, newend)

	return newPrefixIterator(s.prefix, start, end, iter)
}

var _ types.Iterator = (*prefixIterator[[]byte])(nil)

type prefixIterator[V any] struct {
	prefix []byte
	start  []byte
	end    []byte
	iter   types.GIterator[V]
	valid  bool
}

func newPrefixIterator[V any](prefix, start, end []byte, parent types.GIterator[V]) *prefixIterator[V] {
	return &prefixIterator[V]{
		prefix: prefix,
		start:  start,
		end:    end,
		iter:   parent,
		valid:  parent.Valid() && bytes.HasPrefix(parent.Key(), prefix),
	}
}

// Domain implements Iterator, returning the start and end keys of the prefixIterator.
func (pi *prefixIterator[V]) Domain() ([]byte, []byte) {
	return pi.start, pi.end
}

// Valid implements Iterator, checking if the prefixIterator is valid and if the underlying iterator is valid.
func (pi *prefixIterator[V]) Valid() bool {
	return pi.valid && pi.iter.Valid()
}

// Next implements Iterator, moving the underlying iterator to the next key/value pair that starts with the prefix.
func (pi *prefixIterator[V]) Next() {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Next()")
	}

	if pi.iter.Next(); !pi.iter.Valid() || !bytes.HasPrefix(pi.iter.Key(), pi.prefix) {
		// TODO: shouldn't pi be set to nil instead?
		pi.valid = false
	}
}

// Key implements Iterator, returning the stripped prefix key
func (pi *prefixIterator[V]) Key() (key []byte) {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Key()")
	}

	key = pi.iter.Key()
	key = stripPrefix(key, pi.prefix)

	return key
}

// Implements Iterator
func (pi *prefixIterator[V]) Value() V {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Value()")
	}

	return pi.iter.Value()
}

// Close implements Iterator, closing the underlying iterator.
func (pi *prefixIterator[V]) Close() error {
	return pi.iter.Close()
}

// Error returns an error if the prefixIterator is invalid defined by the Valid
// method.
func (pi *prefixIterator[V]) Error() error {
	if !pi.Valid() {
		return errors.New("invalid prefixIterator")
	}

	return nil
}

// stripPrefix is copied from github.com/cometbft/cometbft/libs/db/prefix_db.go
func stripPrefix(key, prefix []byte) []byte {
	if len(key) < len(prefix) || !bytes.Equal(key[:len(prefix)], prefix) {
		panic("should not happen")
	}

	return key[len(prefix):]
}

// cpIncr wraps the bytes in types.PrefixEndBytes
func cpIncr(bz []byte) []byte {
	return types.PrefixEndBytes(bz)
}
