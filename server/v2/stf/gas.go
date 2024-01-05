package stf

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

// Gas consumption descriptors.
const (
	GasDescIterNextCostFlat = "IterNextFlat"
	GasDescValuePerByte     = "ValuePerByte"
	GasDescWritePerByte     = "WritePerByte"
	GasDescReadPerByte      = "ReadPerByte"
	GasDescWriteCostFlat    = "WriteFlat"
	GasDescReadCostFlat     = "ReadFlat"
	GasDescHas              = "Has"
	GasDescDelete           = "Delete"
)

type Store struct {
	parent    store.WritableState
	gasMeter  store.GasMeter
	gasConfig store.GasConfig
}

func New(p store.WritableState, gm store.GasMeter, gc store.GasConfig) store.WritableState {
	return &Store{
		parent:    p,
		gasMeter:  gm,
		gasConfig: gc,
	}
}

func (s *Store) Get(key []byte) ([]byte, error) {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.ReadCostFlat(), GasDescReadCostFlat); err != nil {
		return nil, err
	}

	value, err := s.parent.Get(key)
	if err := s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte()*store.Gas(len(key)), GasDescReadPerByte); err != nil {
		return nil, err
	}
	if err := s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte()*store.Gas(len(value)), GasDescReadPerByte); err != nil {
		return nil, err
	}

	return value, err
}

func (s *Store) Has(key []byte) (bool, error) {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.HasCost(), GasDescHas); err != nil {
		return false, err
	}

	return s.parent.Has(key)
}

func (s *Store) Set(key, value []byte) error {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.WriteCostFlat(), GasDescWriteCostFlat); err != nil {
		return err
	}
	if err := s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte()*store.Gas(len(key)), GasDescWritePerByte); err != nil {
		return err
	}
	if err := s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte()*store.Gas(len(value)), GasDescWritePerByte); err != nil {
		return err
	}

	return s.parent.Set(key, value)
}

func (s *Store) Delete(key []byte) error {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.DeleteCost(), GasDescDelete); err != nil {
		return err
	}

	return s.parent.Delete(key)
}

func (b *Store) ApplyChangeSets(changes []store.ChangeSet) error {
	return b.parent.ApplyChangeSets(changes)
}

func (b *Store) ChangeSets() ([]store.ChangeSet, error) {
	return b.parent.ChangeSets()
}

func (s *Store) Iterator(start, end []byte) (corestore.Iterator, error) {
	itr, err := s.parent.Iterator(start, end)
	if err != nil {
		return nil, err
	}

	return newIterator(itr, s.gasMeter, s.gasConfig), nil
}

func (s *Store) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	itr, err := s.parent.ReverseIterator(start, end)
	if err != nil {
		return nil, err
	}

	return newIterator(itr, s.gasMeter, s.gasConfig), nil
}

var _ corestore.Iterator = (*iterator)(nil)

type iterator struct {
	gasMeter  store.GasMeter
	gasConfig store.GasConfig
	parent    corestore.Iterator
}

func newIterator(parent corestore.Iterator, gm store.GasMeter, gc store.GasConfig) corestore.Iterator {
	return &iterator{
		parent:    parent,
		gasConfig: gc,
		gasMeter:  gm,
	}
}

func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.parent.Domain()
}

func (itr *iterator) Valid() bool {
	return itr.parent.Valid()
}

func (itr *iterator) Key() []byte {
	return itr.parent.Key()
}

func (itr *iterator) Value() []byte {
	return itr.parent.Value()
}

func (itr *iterator) Next() {
	if err := itr.consumeGasSeek(); err != nil {
		// closing the iterator prematurely to prevent further execution
		// TODO: see if this causes any issues
		itr.parent.Close()
		return
	}
	itr.parent.Next()
}

func (itr *iterator) Close() error {
	return itr.parent.Close()
}

func (itr *iterator) Error() error {
	return itr.parent.Error()
}

// consumeGasSeek consumes a fixed amount of gas for each iteration step and a
// variable gas cost based on the current key and value's length. This is called
// prior to the iterator's Next() call.
func (itr *iterator) consumeGasSeek() error {
	if itr.Valid() {
		key := itr.Key()
		value := itr.Value()

		if err := itr.gasMeter.ConsumeGas(itr.gasConfig.ReadCostPerByte()*store.Gas(len(key)), GasDescValuePerByte); err != nil {
			return err
		}
		if err := itr.gasMeter.ConsumeGas(itr.gasConfig.ReadCostPerByte()*store.Gas(len(value)), GasDescValuePerByte); err != nil {
			return err
		}
	}

	if err := itr.gasMeter.ConsumeGas(itr.gasConfig.IterNextCostFlat(), GasDescIterNextCostFlat); err != nil {
		return err
	}

	return nil
}
