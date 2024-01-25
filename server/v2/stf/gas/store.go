package gas

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/stf"
	"cosmossdk.io/server/v2/core/store"
)

// Gas consumption descriptors.
const (
	DescIterNextCostFlat = "IterNextFlat"
	DescValuePerByte     = "ValuePerByte"
	DescWritePerByte     = "WritePerByte"
	DescReadPerByte      = "ReadPerByte"
	DescWriteCostFlat    = "WriteFlat"
	DescReadCostFlat     = "ReadFlat"
	DescHas              = "Has"
	DescDelete           = "Delete"
)

type StoreConfig struct {
	ReadCostFlat, ReadCostPerByte, HasCost          stf.Gas
	WriteCostFlat, WriteCostPerByte, DeleteCostFlat stf.Gas
	IterNextCostFlat                                stf.Gas
}

type Store struct {
	parent    store.Writer
	gasMeter  stf.GasMeter
	gasConfig StoreConfig
}

func New(gc StoreConfig) store.Writer {
	return &Store{
		gasConfig: gc,
	}
}

func (s *Store) Get(key []byte) ([]byte, error) {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.ReadCostFlat, DescReadCostFlat); err != nil {
		return nil, err
	}

	value, err := s.parent.Get(key)
	if err := s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte*stf.Gas(len(key)), DescReadPerByte); err != nil {
		return nil, err
	}
	if err := s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte*stf.Gas(len(value)), DescReadPerByte); err != nil {
		return nil, err
	}

	return value, err
}

func (s *Store) Has(key []byte) (bool, error) {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.HasCost, DescHas); err != nil {
		return false, err
	}

	return s.parent.Has(key)
}

func (s *Store) Set(key, value []byte) error {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.WriteCostFlat, DescWriteCostFlat); err != nil {
		return err
	}
	if err := s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte*stf.Gas(len(key)), DescWritePerByte); err != nil {
		return err
	}
	if err := s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte*stf.Gas(len(value)), DescWritePerByte); err != nil {
		return err
	}

	return s.parent.Set(key, value)
}

func (s *Store) Delete(key []byte) error {
	if err := s.gasMeter.ConsumeGas(s.gasConfig.DeleteCostFlat, DescDelete); err != nil {
		return err
	}

	return s.parent.Delete(key)
}

func (s *Store) ApplyChangeSets(changes []store.KVPair) error {
	return s.parent.ApplyChangeSets(changes)
}

func (s *Store) ChangeSets() ([]store.KVPair, error) {
	return s.parent.ChangeSets()
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
	gasMeter  stf.GasMeter
	gasConfig StoreConfig
	parent    corestore.Iterator
}

func newIterator(parent corestore.Iterator, gm stf.GasMeter, gc StoreConfig) corestore.Iterator {
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

		if err := itr.gasMeter.ConsumeGas(itr.gasConfig.ReadCostPerByte*stf.Gas(len(key)), DescValuePerByte); err != nil {
			return err
		}
		if err := itr.gasMeter.ConsumeGas(itr.gasConfig.ReadCostPerByte*stf.Gas(len(value)), DescValuePerByte); err != nil {
			return err
		}
	}

	if err := itr.gasMeter.ConsumeGas(itr.gasConfig.IterNextCostFlat, DescIterNextCostFlat); err != nil {
		return err
	}

	return nil
}
