package gas

import (
	"io"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.KVStore = &gasKVStore{}

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

// Implements sdk.KVStore.
func (gi *gasKVStore) Get(key []byte) (value []byte) {
	gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostFlat, sdk.GasReadCostFlatDesc)
	value = gi.parent.Get(key)
	// TODO overflow-safe math?
	gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*sdk.Gas(len(value)), sdk.GasReadPerByteDesc)

	return value
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Set(key []byte, value []byte) {
	gi.gasMeter.ConsumeGas(gi.gasConfig.WriteCostFlat, sdk.GasWriteCostFlatDesc)
	// TODO overflow-safe math?
	gi.gasMeter.ConsumeGas(gi.gasConfig.WriteCostPerByte*sdk.Gas(len(value)), sdk.GasWritePerByteDesc)
	gi.parent.Set(key, value)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Has(key []byte) bool {
	gi.gasMeter.ConsumeGas(gi.gasConfig.HasCost, sdk.GasHasDesc)
	return gi.parent.Has(key)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Delete(key []byte) {
	gi.gasMeter.ConsumeGas(gi.gasConfig.DeleteCost, sdk.GasDeleteDesc)
	gi.parent.Delete(key)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) Iterator(start, end []byte) sdk.Iterator {
	return gi.iterator(start, end, true)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) ReverseIterator(start, end []byte) sdk.Iterator {
	return gi.iterator(start, end, false)
}

// Implements sdk.KVStore.
func (gi *gasKVStore) CacheWrap() sdk.CacheWrap {
	panic("cannot CacheWrap a GasKVStore")
}

// CacheWrapWithTrace implements the sdk.KVStore interface.
func (gi *gasKVStore) CacheWrapWithTrace(_ io.Writer, _ sdk.TraceContext) sdk.CacheWrap {
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
