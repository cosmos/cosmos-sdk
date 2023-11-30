package gaskv

import (
	"fmt"
	"io"
	"time"

	"github.com/streadway/handy/atomic"

	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
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

// Implements Store.
func (gs *Store) GetStoreType() types.StoreType {
	return gs.parent.GetStoreType()
}

var DebugLogging atomic.Int

var gasTotal uint64

func StartLogging() {
	DebugLogging.Add(1)
}

func StopLogging() {
	DebugLogging.Add(-1)
	fmt.Printf("GasKVStore DebugLog total=%d\n", gasTotal)
	gasTotal = 0
}

func debugLog(op string, sub string, amount uint64) {
	if DebugLogging.Get() == 1 {
		fmt.Printf("GasKVStore DebugLog op=%s sub=%s amount=%d\n", op, sub, amount)
		gasTotal += amount
	}
}

// Implements KVStore.
func (gs *Store) Get(key []byte) (value []byte) {
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostFlat, types.GasReadCostFlatDesc)
	value = gs.parent.Get(key)
	debugLog("Get", "Flat", gs.gasConfig.ReadCostFlat)

	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*types.Gas(len(key)), types.GasReadPerByteDesc)
	debugLog("Get", "PerByteKey", gs.gasConfig.ReadCostPerByte*types.Gas(len(key)))
	if DebugLogging.Get() == 1 {
		fmt.Printf("GasKVStore DebugLog Get %d*%d=%d len(key)=%d\n",
			gs.gasConfig.ReadCostPerByte,
			types.Gas(len(key)),
			gs.gasConfig.ReadCostPerByte*types.Gas(len(key)),
			len(key),
		)
		if gs.gasConfig.ReadCostPerByte*types.Gas(len(key)) == 63 || gs.gasConfig.ReadCostPerByte*types.Gas(len(key)) == 60 {
			fmt.Printf("GasKVStore DebugLog key=%s\n", key)
			if gs.gasConfig.ReadCostPerByte*types.Gas(len(value)) == 147 {
				fmt.Println("found it")
			}
		}
	}
	gs.gasMeter.ConsumeGas(gs.gasConfig.ReadCostPerByte*types.Gas(len(value)), types.GasReadPerByteDesc)
	debugLog("Get", "PerByteValue", gs.gasConfig.ReadCostPerByte*types.Gas(len(value)))

	return value
}

// Implements KVStore.
func (gs *Store) Set(key []byte, value []byte) {
	types.AssertValidKey(key)
	types.AssertValidValue(value)
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostFlat, types.GasWriteCostFlatDesc)
	debugLog("Set", "Flat", gs.gasConfig.WriteCostFlat)
	// TODO overflow-safe math?
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*types.Gas(len(key)), types.GasWritePerByteDesc)
	debugLog("Set", "PerByteKey", gs.gasConfig.WriteCostPerByte*types.Gas(len(key)))
	gs.gasMeter.ConsumeGas(gs.gasConfig.WriteCostPerByte*types.Gas(len(value)), types.GasWritePerByteDesc)
	debugLog("Set", "PerByteValue", gs.gasConfig.WriteCostPerByte*types.Gas(len(value)))
	gs.parent.Set(key, value)
}

// Implements KVStore.
func (gs *Store) Has(key []byte) bool {
	defer telemetry.MeasureSince(time.Now(), "store", "gaskv", "has")
	gs.gasMeter.ConsumeGas(gs.gasConfig.HasCost, types.GasHasDesc)
	debugLog("Has", "Flat", gs.gasConfig.HasCost)
	return gs.parent.Has(key)
}

// Implements KVStore.
func (gs *Store) Delete(key []byte) {
	defer telemetry.MeasureSince(time.Now(), "store", "gaskv", "delete")
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

// Implements KVStore.
func (gs *Store) CacheWrap() types.CacheWrap {
	panic("cannot CacheWrap a GasKVStore")
}

// CacheWrapWithTrace implements the KVStore interface.
func (gs *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	panic("cannot CacheWrapWithTrace a GasKVStore")
}

// CacheWrapWithListeners implements the CacheWrapper interface.
func (gs *Store) CacheWrapWithListeners(_ types.StoreKey, _ []types.WriteListener) types.CacheWrap {
	panic("cannot CacheWrapWithListeners a GasKVStore")
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
	defer telemetry.MeasureSince(time.Now(), "store", "gaskv", "iterator", "next")
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

// Implements Iterator.
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

		gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*types.Gas(len(key)), types.GasValuePerByteDesc)
		debugLog("consumeSeekGas", "PerByteKey", gi.gasConfig.ReadCostPerByte*types.Gas(len(key)))
		gi.gasMeter.ConsumeGas(gi.gasConfig.ReadCostPerByte*types.Gas(len(value)), types.GasValuePerByteDesc)
		debugLog("consumeSeekGas", "PerByteValue", gi.gasConfig.ReadCostPerByte*types.Gas(len(value)))
	}

	gi.gasMeter.ConsumeGas(gi.gasConfig.IterNextCostFlat, types.GasIterNextCostFlatDesc)
	debugLog("consumeSeekGas", "Flat", gi.gasConfig.IterNextCostFlat)
}
