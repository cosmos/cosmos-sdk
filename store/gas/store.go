package gas

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.KVStore = &gasKVStore{}

// gasKVStore applies gas tracking to an underlying KVStore. It implements the
// KVStore interface.
type gasKVStore struct {
	tank   *sdk.GasTank
	parent sdk.KVStore
}

// NewGasKVStore returns a reference to a new GasKVStore.
// nolint
func NewStore(tank *sdk.GasTank, parent sdk.KVStore) *gasKVStore {
	kvs := &gasKVStore{
		tank:   tank,
		parent: parent,
	}
	return kvs
}

// Implements sdk.KVStore.
func (gs *gasKVStore) Get(key []byte) (value []byte) {
	gs.tank.ReadFlat()
	value = gs.parent.Get(key)
	// TODO overflow-safe math?
	gs.tank.ReadBytes(len(value))

	return value
}

// Implements sdk.KVStore.
func (gs *gasKVStore) Set(key []byte, value []byte) {
	gs.tank.WriteFlat()
	// TODO overflow-safe math?
	gs.tank.WriteBytes(len(value))
	gs.parent.Set(key, value)
}

// Implements sdk.KVStore.
func (gs *gasKVStore) Has(key []byte) bool {
	gs.tank.HasFlat()
	return gs.parent.Has(key)
}

// Implements sdk.KVStore.
func (gs *gasKVStore) Delete(key []byte) {
	gs.tank.DeleteFlat()
	gs.parent.Delete(key)
}

// Implements sdk.KVStore.
func (gs *gasKVStore) Iterator(start, end []byte) sdk.Iterator {
	return gs.iterator(start, end, true)
}

// Implements sdk.KVStore.
func (gs *gasKVStore) ReverseIterator(start, end []byte) sdk.Iterator {
	return gs.iterator(start, end, false)
}

func (gs *gasKVStore) iterator(start, end []byte, ascending bool) sdk.Iterator {
	var parent sdk.Iterator
	if ascending {
		parent = gs.parent.Iterator(start, end)
	} else {
		parent = gs.parent.ReverseIterator(start, end)
	}

	gi := newGasIterator(gs.tank, parent)
	if gi.Valid() {
		gi.(*gasIterator).consumeSeekGas()
	}

	return gi
}

type gasIterator struct {
	tank   *sdk.GasTank
	parent sdk.Iterator
}

func newGasIterator(tank *sdk.GasTank, parent sdk.Iterator) sdk.Iterator {
	return &gasIterator{
		tank:   tank,
		parent: parent,
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

	gi.tank.ValueBytes(len(value))
	gi.tank.IterNextFlat()
}
