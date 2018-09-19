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
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, sdk.GasReadFlatDesc)
	value = gs.parent.Get(key)

	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*sdk.Gas(len(value)), sdk.GasReadPerByteDesc)

	return value
}

// Implements KVStore.
func (gs *gasKVStore) Set(key []byte, value []byte) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, sdk.GasWriteFlatDesc)
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
	// No gas costs for deletion
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

// Implements KVStore.
func (gs *gasKVStore) Iterator(start, end []byte) sdk.Iterator {
	return gs.iterator(start, end, true)
}

// Implements KVStore.
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

	gs.gasMeter.ConsumeGas(gs.gasConfig.IterInitFlat, sdk.GasIterInitFlatDesc)
	return newGasIterator(gs.gasMeter, gs.gasConfig, parent)
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

// Implements Iterator.
func (gi *gasIterator) Next() {
	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextFlat, sdk.GasIterNextFlatDesc)
	gi.parent.Next()
}

// Implements Iterator.
func (gi *gasIterator) Key() (key []byte) {
	key = gi.parent.Key()
	return key
}

// Implements Iterator.
func (gi *gasIterator) Value() (value []byte) {
	value = gi.parent.Value()
	gi.gasMeter.ConsumeGas(gi.gasConfig.ValueCostPerByte*sdk.Gas(len(value)), sdk.GasValuePerByteDesc)
	return value
}

// Implements Iterator.
func (gi *gasIterator) Close() {
	gi.parent.Close()
}
