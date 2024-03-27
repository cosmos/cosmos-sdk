package gaskv

import (
	"io"

	"cosmossdk.io/store/types"
)

// ObjectValueLength is the emulated number of bytes for storing transient objects in gas accounting.
const ObjectValueLength = 16

var _ types.KVStore = &Store{}

type Store = GStore[[]byte]

func NewStore(parent types.KVStore, gasMeter types.GasMeter, gasConfig types.GasConfig) *Store {
	return NewGStore(parent, gasMeter, gasConfig,
		func(v []byte) bool { return v == nil },
		func(v []byte) int { return len(v) },
	)
}

type ObjStore = GStore[any]

func NewObjStore(parent types.ObjKVStore, gasMeter types.GasMeter, gasConfig types.GasConfig) *ObjStore {
	return NewGStore(parent, gasMeter, gasConfig,
		func(v any) bool { return v == nil },
		func(v any) int { return ObjectValueLength },
	)
}

// GStore applies gas tracking to an underlying KVStore. It implements the
// KVStore interface.
type GStore[V any] struct {
	gasMeter  types.GasMeter
	gasConfig types.GasConfig
	parent    types.GKVStore[V]

	isZero   func(V) bool
	valueLen func(V) int
}

// NewGStore returns a reference to a new GasKVStore.
func NewGStore[V any](
	parent types.GKVStore[V],
	gasMeter types.GasMeter,
	gasConfig types.GasConfig,
	isZero func(V) bool,
	valueLen func(V) int,
) *GStore[V] {
	kvs := &GStore[V]{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
		isZero:    isZero,
		valueLen:  valueLen,
	}
	return kvs
}

// GetStoreType implements Store.
func (gs *GStore[V]) GetStoreType() types.StoreType {
	return gs.parent.GetStoreType()
}

// Get implements KVStore.
func (gs *GStore[V]) Get(key []byte) (value V) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, types.GasReadCostFlatDesc)
	value = gs.parent.Get(key)

	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*types.Gas(len(key)), types.GasReadPerByteDesc)
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*types.Gas(gs.valueLen(value)), types.GasReadPerByteDesc)

	return value
}

// Set implements KVStore.
func (gs *GStore[V]) Set(key []byte, value V) {
	types.AssertValidKey(key)
	types.AssertValidValueGeneric(value, gs.isZero, gs.valueLen)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, types.GasWriteCostFlatDesc)
	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*types.Gas(len(key)), types.GasWritePerByteDesc)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*types.Gas(gs.valueLen(value)), types.GasWritePerByteDesc)
	gs.parent.Set(key, value)
}

// Has implements KVStore.
func (gs *GStore[V]) Has(key []byte) bool {
	gs.gasMeter.ConsumeGas(gs.gasConfig.HasCost, types.GasHasDesc)
	return gs.parent.Has(key)
}

// Delete implements KVStore.
func (gs *GStore[V]) Delete(key []byte) {
	// charge gas to prevent certain attack vectors even though space is being freed
	gs.gasMeter.ConsumeGas(gs.gasConfig.DeleteCost, types.GasDeleteDesc)
	gs.parent.Delete(key)
}

// Iterator implements the KVStore interface. It returns an iterator which
// incurs a flat gas cost for seeking to the first key/value pair and a variable
// gas cost based on the current value's length if the iterator is valid.
func (gs *GStore[V]) Iterator(start, end []byte) types.GIterator[V] {
	return gs.iterator(start, end, true)
}

// ReverseIterator implements the KVStore interface. It returns a reverse
// iterator which incurs a flat gas cost for seeking to the first key/value pair
// and a variable gas cost based on the current value's length if the iterator
// is valid.
func (gs *GStore[V]) ReverseIterator(start, end []byte) types.GIterator[V] {
	return gs.iterator(start, end, false)
}

// CacheWrap implements KVStore.
func (gs *GStore[V]) CacheWrap() types.CacheWrap {
	panic("cannot CacheWrap a GasKVStore")
}

// CacheWrapWithTrace implements the KVStore interface.
func (gs *GStore[V]) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	panic("cannot CacheWrapWithTrace a GasKVStore")
}

func (gs *GStore[V]) iterator(start, end []byte, ascending bool) types.GIterator[V] {
	var parent types.GIterator[V]
	if ascending {
		parent = gs.parent.Iterator(start, end)
	} else {
		parent = gs.parent.ReverseIterator(start, end)
	}

	gi := newGasIterator(gs.gasMeter, gs.gasConfig, parent, gs.valueLen)
	gi.consumeSeekGas()

	return gi
}

type gasIterator[V any] struct {
	gasMeter  types.GasMeter
	gasConfig types.GasConfig
	parent    types.GIterator[V]
	valueLen  func(V) int
}

func newGasIterator[V any](gasMeter types.GasMeter, gasConfig types.GasConfig, parent types.GIterator[V], valueLen func(V) int) *gasIterator[V] {
	return &gasIterator[V]{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
		valueLen:  valueLen,
	}
}

// Domain implements Iterator.
func (gi *gasIterator[V]) Domain() (start, end []byte) {
	return gi.parent.Domain()
}

// Valid implements Iterator.
func (gi *gasIterator[V]) Valid() bool {
	return gi.parent.Valid()
}

// Next implements the Iterator interface. It seeks to the next key/value pair
// in the iterator. It incurs a flat gas cost for seeking and a variable gas
// cost based on the current value's length if the iterator is valid.
func (gi *gasIterator[V]) Next() {
	gi.consumeSeekGas()
	gi.parent.Next()
}

// Key implements the Iterator interface. It returns the current key and it does
// not incur any gas cost.
func (gi *gasIterator[V]) Key() (key []byte) {
	key = gi.parent.Key()
	return key
}

// Value implements the Iterator interface. It returns the current value and it
// does not incur any gas cost.
func (gi *gasIterator[V]) Value() (value V) {
	return gi.parent.Value()
}

// Close implements Iterator.
func (gi *gasIterator[V]) Close() error {
	return gi.parent.Close()
}

// Error delegates the Error call to the parent iterator.
func (gi *gasIterator[V]) Error() error {
	return gi.parent.Error()
}

// consumeSeekGas consumes on each iteration step a flat gas cost and a variable gas cost
// based on the current value's length.
func (gi *gasIterator[V]) consumeSeekGas() {
	if gi.Valid() {
		key := gi.Key()
		value := gi.Value()

		gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*types.Gas(len(key)), types.GasValuePerByteDesc)
		gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*types.Gas(gi.valueLen(value)), types.GasValuePerByteDesc)
	}
	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextCostFlat, types.GasIterNextCostFlatDesc)
}
