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
		func(v []byte) bool { return v == nil },
		func(v []byte) int { return len(v) },
	)
}

func NewObjStore(parent types.ObjKVStore, prefix []byte) ObjStore {
	return NewGStore(
		parent, prefix,
		func(v any) bool { return v == nil },
		func(v any) int { return 1 },
	)
}

// GStore is similar with cometbft/cometbft-db/blob/v1.0.1/prefixdb.go
// both gives access only to the limited subset of the store
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
	return
}

func (s GStore[V]) key(key []byte) (res []byte) {
	if key == nil {
		panic("nil key on Store")
	}
	res = cloneAppend(s.prefix, key)
	return
}

// GetStoreType implements Store
func (s GStore[V]) GetStoreType() types.StoreType {
	return s.parent.GetStoreType()
}

// CacheWrap implements CacheWrap
func (s GStore[V]) CacheWrap() types.CacheWrap {
	return cachekv.NewGStore(s, s.isZero, s.valueLen)
}

// CacheWrapWithTrace implements the KVStore interface.
func (s GStore[V]) CacheWrapWithTrace(w io.Writer, tc types.TraceContext) types.CacheWrap {
	if store, ok := any(s).(*GStore[[]byte]); ok {
		return cachekv.NewGStore(tracekv.NewStore(store, w, tc), store.isZero, store.valueLen)
	}
	return s.CacheWrap()
}

// Get implements KVStore
func (s GStore[V]) Get(key []byte) V {
	res := s.parent.Get(s.key(key))
	return res
}

// Has implements KVStore
func (s GStore[V]) Has(key []byte) bool {
	return s.parent.Has(s.key(key))
}

// Set implements KVStore
func (s GStore[V]) Set(key []byte, value V) {
	types.AssertValidKey(key)
	types.AssertValidValueGeneric(value, s.isZero, s.valueLen)
	s.parent.Set(s.key(key), value)
}

// Delete implements KVStore
func (s GStore[V]) Delete(key []byte) {
	s.parent.Delete(s.key(key))
}

// Iterator implements KVStore
// Check https://github.com/cometbft/cometbft-db/blob/v1.0.1/prefixdb.go#L109
func (s GStore[V]) Iterator(start, end []byte) types.GIterator[V] {
	newstart := cloneAppend(s.prefix, start)

	var newend []byte
	if end == nil {
		newend = cpIncr(s.prefix)
	} else {
		newend = cloneAppend(s.prefix, end)
	}

	iter := s.parent.Iterator(newstart, newend)

	return newPrefixIterator(s.prefix, start, end, iter)
}

// ReverseIterator implements KVStore
// Check https://github.com/cometbft/cometbft-db/blob/v1.0.1/prefixdb.go#L132
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

// Domain implements Iterator
func (pi *prefixIterator[V]) Domain() ([]byte, []byte) {
	return pi.start, pi.end
}

// Valid implements Iterator
func (pi *prefixIterator[V]) Valid() bool {
	return pi.valid && pi.iter.Valid()
}

// Next implements Iterator
func (pi *prefixIterator[V]) Next() {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Next()")
	}

	if pi.iter.Next(); !pi.iter.Valid() || !bytes.HasPrefix(pi.iter.Key(), pi.prefix) {
		// TODO: shouldn't pi be set to nil instead?
		pi.valid = false
	}
}

// Key implements Iterator
func (pi *prefixIterator[V]) Key() (key []byte) {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Key()")
	}

	key = pi.iter.Key()
	key = stripPrefix(key, pi.prefix)

	return
}

// Value implements Iterator
func (pi *prefixIterator[V]) Value() V {
	if !pi.valid {
		panic("prefixIterator invalid, cannot call Value()")
	}

	return pi.iter.Value()
}

// Close implements Iterator
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

// copied from github.com/cometbft/cometbft-db/blob/v1.0.1/prefixdb.go
func stripPrefix(key, prefix []byte) []byte {
	if len(key) < len(prefix) || !bytes.Equal(key[:len(prefix)], prefix) {
		panic("should not happen")
	}

	return key[len(prefix):]
}

// wrapping types.PrefixEndBytes
func cpIncr(bz []byte) []byte {
	return types.PrefixEndBytes(bz)
}
