package store

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ KVStore = &gasKVStore{}

// gasKVStore applies gas tracking to an underlying KVStore. It implements the
// KVStore interface.
type gasKVStore struct {
	gasMeter  sdk.GasMeter
	gasConfig sdk.GasConfig
	parent    sdk.KVStore
}

// NewGasKVStore returns a reference to a new GasKVStore.
// nolint
func NewGasKVStore(gasMeter sdk.GasMeter, gasConfig sdk.GasConfig, parent sdk.KVStore) *gasKVStore {
	kvs := &gasKVStore{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
	return kvs
}

// Implements Store.
func (gs *gasKVStore) GetStoreType() sdk.StoreType {
	return gs.parent.GetStoreType()
}

// Implements KVStore.
func (gs *gasKVStore) Get(key []byte) (value []byte) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, sdk.GasReadCostFlatDesc)
	value = gs.parent.Get(key)

	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*sdk.Gas(len(value)), sdk.GasReadPerByteDesc)

	return value
}

// Implements KVStore.
func (gs *gasKVStore) Set(key []byte, value []byte) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, sdk.GasWriteCostFlatDesc)
	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*sdk.Gas(len(value)), sdk.GasWritePerByteDesc)
	gs.parent.Set(key, value)
}

// Implements KVStore.
func (gs *gasKVStore) Has(key []byte) bool {
	gs.gasMeter.ConsumeGas(gs.gasConfig.HasCost, sdk.GasHasDesc)
	return gs.parent.Has(key)
}

// Implements KVStore.
func (gs *gasKVStore) Delete(key []byte) {
	// charge gas to prevent certain attack vectors even though space is being freed
	gs.gasMeter.ConsumeGas(gs.gasConfig.DeleteCost, sdk.GasDeleteDesc)
	gs.parent.Delete(key)
}

// Implements KVStore
func (gs *gasKVStore) Prefix(prefix []byte) KVStore {
	// Keep gasstore layer at the top
	return &gasKVStore{
		gasMeter:  gs.gasMeter,
		gasConfig: gs.gasConfig,
		parent:    prefixStore{gs.parent, prefix},
	}
}

// Implements KVStore
func (gs *gasKVStore) Gas(meter GasMeter, config GasConfig) KVStore {
	return NewGasKVStore(meter, config, gs)
}

// Iterator implements the KVStore interface. It returns an iterator which
// incurs a flat gas cost for seeking to the first key/value pair and a variable
// gas cost based on the current value's length if the iterator is valid.
func (gs *gasKVStore) Iterator(start, end []byte) sdk.Iterator {
	return gs.iterator(start, end, true)
}

// ReverseIterator implements the KVStore interface. It returns a reverse
// iterator which incurs a flat gas cost for seeking to the first key/value pair
// and a variable gas cost based on the current value's length if the iterator
// is valid.
func (gs *gasKVStore) ReverseIterator(start, end []byte) sdk.Iterator {
	return gs.iterator(start, end, false)
}

// Implements KVStore.
func (gs *gasKVStore) CacheWrap() sdk.CacheWrap {
	panic("cannot CacheWrap a GasKVStore")
}

// CacheWrapWithTrace implements the KVStore interface.
func (gs *gasKVStore) CacheWrapWithTrace(_ io.Writer, _ TraceContext) CacheWrap {
	panic("cannot CacheWrapWithTrace a GasKVStore")
}

func (gs *gasKVStore) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator
	if ascending {
		parent = gs.parent.Iterator(start, end)
	} else {
		parent = gs.parent.ReverseIterator(start, end)
	}

	gi := newGasIterator(gs.gasMeter, gs.gasConfig, parent)
	if gi.Valid() {
		gi.(*gasIterator).consumeSeekGas()
	}

	return gi
}

type gasIterator struct {
	gasMeter  sdk.GasMeter
	gasConfig sdk.GasConfig
	parent    sdk.Iterator
}

func newGasIterator(gasMeter sdk.GasMeter, gasConfig sdk.GasConfig, parent sdk.Iterator) sdk.Iterator {
	return &gasIterator{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
}

// Implements Iterator.
func (gi *gasIterator) Domain() (start []byte, end []byte) {
	return gi.parent.Domain()
}

// Implements Iterator.
func (gi *gasIterator) Valid() bool {
	return gi.parent.Valid()
}

// Next implements the Iterator interface. It seeks to the next key/value pair
// in the iterator. It incurs a flat gas cost for seeking and a variable gas
// cost based on the current value's length if the iterator is valid.
func (gi *gasIterator) Next() {
	if gi.Valid() {
		gi.consumeSeekGas()
	}

	gi.parent.Next()
}

// Key implements the Iterator interface. It returns the current key and it does
// not incur any gas cost.
func (gi *gasIterator) Key() (key []byte) {
	key = gi.parent.Key()
	return key
}

// Value implements the Iterator interface. It returns the current value and it
// does not incur any gas cost.
func (gi *gasIterator) Value() (value []byte) {
	value = gi.parent.Value()
	return value
}

// Implements Iterator.
func (gi *gasIterator) Close() {
	gi.parent.Close()
}

// consumeSeekGas consumes a flat gas cost for seeking and a variable gas cost
// based on the current value's length.
func (gi *gasIterator) consumeSeekGas() {
	value := gi.Value()

	gi.gasMeter.ConsumeGas(gi.gasConfig.ValueCostPerByte*sdk.Gas(len(value)), sdk.GasValuePerByteDesc)
	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextCostFlat, sdk.GasIterNextCostFlatDesc)
}
