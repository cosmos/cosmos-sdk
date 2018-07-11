package store

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint
const (
	HasCost          = 10
	ReadCostFlat     = 10
	ReadCostPerByte  = 1
	WriteCostFlat    = 10
	WriteCostPerByte = 10
	KeyCostFlat      = 5
	ValueCostFlat    = 10
	ValueCostPerByte = 1
)

// gasKVStore applies gas tracking to an underlying kvstore
type gasKVStore struct {
	gasMeter sdk.GasMeter
	parent   sdk.KVStore
}

// nolint
func NewGasKVStore(gasMeter sdk.GasMeter, parent sdk.KVStore) *gasKVStore {
	kvs := &gasKVStore{
		gasMeter: gasMeter,
		parent:   parent,
	}
	return kvs
}

// Implements Store.
func (gi *gasKVStore) GetStoreType() sdk.StoreType {
	return gi.parent.GetStoreType()
}

// Implements KVStore.
func (gi *gasKVStore) Get(key []byte) (value []byte) {
	gi.gasMeter.ConsumeGas(ReadCostFlat, "GetFlat")
	value = gi.parent.Get(key)
	// TODO overflow-safe math?
	gi.gasMeter.ConsumeGas(ReadCostPerByte*sdk.Gas(len(value)), "ReadPerByte")
	return value
}

// Implements KVStore.
func (gi *gasKVStore) Set(key []byte, value []byte) {
	gi.gasMeter.ConsumeGas(WriteCostFlat, "SetFlat")
	// TODO overflow-safe math?
	gi.gasMeter.ConsumeGas(WriteCostPerByte*sdk.Gas(len(value)), "SetPerByte")
	gi.parent.Set(key, value)
}

// Implements KVStore.
func (gi *gasKVStore) Has(key []byte) bool {
	gi.gasMeter.ConsumeGas(HasCost, "Has")
	return gi.parent.Has(key)
}

// Implements KVStore.
func (gi *gasKVStore) Delete(key []byte) {
	// No gas costs for deletion
	gi.parent.Delete(key)
}

// Implements KVStore
func (gi *gasKVStore) Prefix(prefix []byte) KVStore {
	return prefixStore{gi, prefix}
}

// Implements KVStore.
func (gi *gasKVStore) Iterator(start, end []byte) sdk.Iterator {
	return gi.iterator(start, end, true)
}

// Implements KVStore.
func (gi *gasKVStore) ReverseIterator(start, end []byte) sdk.Iterator {
	return gi.iterator(start, end, false)
}

// Implements KVStore.
func (gi *gasKVStore) CacheWrap() sdk.CacheWrap {
	panic("you cannot CacheWrap a GasKVStore")
}

func (gi *gasKVStore) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator
	if ascending {
		parent = gi.parent.Iterator(start, end)
	} else {
		parent = gi.parent.ReverseIterator(start, end)
	}
	return newGasIterator(gi.gasMeter, parent)
}

type gasIterator struct {
	gasMeter sdk.GasMeter
	parent   sdk.Iterator
}

func newGasIterator(gasMeter sdk.GasMeter, parent sdk.Iterator) sdk.Iterator {
	return &gasIterator{
		gasMeter: gasMeter,
		parent:   parent,
	}
}

// Implements Iterator.
func (g *gasIterator) Domain() (start []byte, end []byte) {
	return g.parent.Domain()
}

// Implements Iterator.
func (g *gasIterator) Valid() bool {
	return g.parent.Valid()
}

// Implements Iterator.
func (g *gasIterator) Next() {
	g.parent.Next()
}

// Implements Iterator.
func (g *gasIterator) Key() (key []byte) {
	g.gasMeter.ConsumeGas(KeyCostFlat, "KeyFlat")
	key = g.parent.Key()
	return key
}

// Implements Iterator.
func (g *gasIterator) Value() (value []byte) {
	value = g.parent.Value()
	g.gasMeter.ConsumeGas(ValueCostFlat, "ValueFlat")
	g.gasMeter.ConsumeGas(ValueCostPerByte*sdk.Gas(len(value)), "ValuePerByte")
	return value
}

// Implements Iterator.
func (g *gasIterator) Close() {
	g.parent.Close()
}
