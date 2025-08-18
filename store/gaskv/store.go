package gaskv

import (
	"fmt"
	"io"

	"cosmossdk.io/store/types"
)

var _ types.KVStore = &Store{}

// Store applies gas tracking to an underlying KVStore. It implements the
// KVStore interface.
type Store struct {
	gasMeter  types.GasMeter
	gasConfig types.GasConfig
	parent    types.KVStore
}

// NewStore returns a reference to a new GasKVStore.
func NewStore(parent types.KVStore, gasMeter types.GasMeter, gasConfig types.GasConfig) *Store {
	kvs := &Store{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
	return kvs
}

// GetStoreType implements Store, consuming no gas and returning the underlying
// store's type.
func (gs *Store) GetStoreType() types.StoreType {
	return gs.parent.GetStoreType()
}

// Get implements KVStore, consuming gas based on ReadCostFlat and the read per bytes cost.
func (gs *Store) Get(key []byte) (value []byte) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, types.GasReadCostFlatDesc)
	value = gs.parent.Get(key)

	// Safe gas calculation for key length
	if gasCost, err := SafeMul(gs.gasConfig.ReadCostPerByte, len(key)); err == nil {
		gs.gasMeter.ConsumeGas(gasCost, types.GasReadPerByteDesc)
	} else {
		// If overflow occurs, trigger out-of-gas panic deterministically
		remaining := gs.gasMeter.Limit() - gs.gasMeter.GasConsumed()
		gs.gasMeter.ConsumeGas(remaining+1, types.GasReadPerByteDesc)
	}

	// Safe gas calculation for value length
	if gasCost, err := SafeMul(gs.gasConfig.ReadCostPerByte, len(value)); err == nil {
		gs.gasMeter.ConsumeGas(gasCost, types.GasReadPerByteDesc)
	} else {
		// If overflow occurs, trigger out-of-gas panic deterministically
		remaining := gs.gasMeter.Limit() - gs.gasMeter.GasConsumed()
		gs.gasMeter.ConsumeGas(remaining+1, types.GasReadPerByteDesc)
	}

	return value
}

// Set implements KVStore, consuming gas based on WriteCostFlat and the write per bytes cost.
func (gs *Store) Set(key, value []byte) {
	types.AssertValidKey(key)
	types.AssertValidValue(value)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, types.GasWriteCostFlatDesc)

	// Safe gas calculation for key length
	if gasCost, err := SafeMul(gs.gasConfig.WriteCostPerByte, len(key)); err == nil {
		gs.gasMeter.ConsumeGas(gasCost, types.GasWritePerByteDesc)
	} else {
		// If overflow occurs, trigger out-of-gas panic deterministically
		remaining := gs.gasMeter.Limit() - gs.gasMeter.GasConsumed()
		gs.gasMeter.ConsumeGas(remaining+1, types.GasWritePerByteDesc)
	}

	// Safe gas calculation for value length
	if gasCost, err := SafeMul(gs.gasConfig.WriteCostPerByte, len(value)); err == nil {
		gs.gasMeter.ConsumeGas(gasCost, types.GasWritePerByteDesc)
	} else {
		// If overflow occurs, trigger out-of-gas panic deterministically
		remaining := gs.gasMeter.Limit() - gs.gasMeter.GasConsumed()
		gs.gasMeter.ConsumeGas(remaining+1, types.GasWritePerByteDesc)
	}

	gs.parent.Set(key, value)
}

// Has implements KVStore, consuming gas based on HasCost.
func (gs *Store) Has(key []byte) bool {
	gs.gasMeter.ConsumeGas(gs.gasConfig.HasCost, types.GasHasDesc)
	return gs.parent.Has(key)
}

// Delete implements KVStore consuming gas based on DeleteCost.
func (gs *Store) Delete(key []byte) {
	// charge gas to prevent certain attack vectors even though space is being freed
	gs.gasMeter.ConsumeGas(gs.gasConfig.DeleteCost, types.GasDeleteDesc)
	gs.parent.Delete(key)
}

// Iterator implements the KVStore interface. It returns an iterator which
// incurs a flat gas cost for seeking to the first key/value pair and a variable
// gas cost based on the current value's length if the iterator is valid.
func (gs *Store) Iterator(start, end []byte) types.Iterator {
	return gs.iterator(start, end, true)
}

// ReverseIterator implements the KVStore interface. It returns a reverse
// iterator which incurs a flat gas cost for seeking to the first key/value pair
// and a variable gas cost based on the current value's length if the iterator
// is valid.
func (gs *Store) ReverseIterator(start, end []byte) types.Iterator {
	return gs.iterator(start, end, false)
}

// CacheWrap implements KVStore - it PANICS as you cannot cache a GasKVStore.
func (gs *Store) CacheWrap() types.CacheWrap {
	panic("cannot CacheWrap a GasKVStore")
}

// CacheWrapWithTrace implements the KVStore interface.
func (gs *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	panic("cannot CacheWrapWithTrace a GasKVStore")
}

func (gs *Store) iterator(start, end []byte, ascending bool) types.Iterator {
	var parent types.Iterator
	if ascending {
		parent = gs.parent.Iterator(start, end)
	} else {
		parent = gs.parent.ReverseIterator(start, end)
	}

	gi := newGasIterator(gs.gasMeter, gs.gasConfig, parent)
	gi.(*gasIterator).consumeSeekGas()

	return gi
}

type gasIterator struct {
	gasMeter  types.GasMeter
	gasConfig types.GasConfig
	parent    types.Iterator
}

func newGasIterator(gasMeter types.GasMeter, gasConfig types.GasConfig, parent types.Iterator) types.Iterator {
	return &gasIterator{
		gasMeter:  gasMeter,
		gasConfig: gasConfig,
		parent:    parent,
	}
}

// Domain implements Iterator, getting the underlying iterator's domain.
func (gi *gasIterator) Domain() (start, end []byte) {
	return gi.parent.Domain()
}

// Valid implements Iterator by checking the underlying iterator.
func (gi *gasIterator) Valid() bool {
	return gi.parent.Valid()
}

// Next implements the Iterator interface. It seeks to the next key/value pair
// in the iterator. It incurs a flat gas cost for seeking and a variable gas
// cost based on the current value's length if the iterator is valid.
func (gi *gasIterator) Next() {
	gi.consumeSeekGas()
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

// Close implements Iterator by closing the underlying iterator.
func (gi *gasIterator) Close() error {
	return gi.parent.Close()
}

// Error delegates the Error call to the parent iterator.
func (gi *gasIterator) Error() error {
	return gi.parent.Error()
}

// consumeSeekGas consumes on each iteration step a flat gas cost and a variable gas cost
// based on the current value's length.
func (gi *gasIterator) consumeSeekGas() {
	if gi.Valid() {
		key := gi.Key()
		value := gi.Value()

		// Safe gas calculation for key length
		if gasCost, err := SafeMul(gi.gasConfig.ReadCostPerByte, len(key)); err == nil {
			gi.gasMeter.ConsumeGas(gasCost, types.GasValuePerByteDesc)
		} else {
			// If overflow occurs, trigger out-of-gas panic deterministically
			remaining := gi.gasMeter.Limit() - gi.gasMeter.GasConsumed()
			gi.gasMeter.ConsumeGas(remaining+1, types.GasValuePerByteDesc)
		}

		// Safe gas calculation for value length
		if gasCost, err := SafeMul(gi.gasConfig.ReadCostPerByte, len(value)); err == nil {
			gi.gasMeter.ConsumeGas(gasCost, types.GasValuePerByteDesc)
		} else {
			// If overflow occurs, trigger out-of-gas panic deterministically
			remaining := gi.gasMeter.Limit() - gi.gasMeter.GasConsumed()
			gi.gasMeter.ConsumeGas(remaining+1, types.GasValuePerByteDesc)
		}
	}
	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextCostFlat, types.GasIterNextCostFlatDesc)
}

// SafeMul performs safe multiplication of gas cost and length to prevent overflow
func SafeMul(cost types.Gas, length int) (types.Gas, error) {
	if length < 0 {
		return 0, fmt.Errorf("negative length: %d", length)
	}
	if cost == 0 {
		return 0, nil
	}

	// Check for overflow: if cost * uint64(length) would overflow uint64
	if uint64(length) > 0 && cost > types.Gas(^uint64(0))/types.Gas(length) {
		return 0, fmt.Errorf("gas calculation overflow: cost=%d, length=%d", cost, length)
	}

	return cost * types.Gas(length), nil
}
