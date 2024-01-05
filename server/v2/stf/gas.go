package stf

import (
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
	s.gasMeter.ConsumeGas(s.gasConfig.ReadCostFlat(), GasDescReadCostFlat)

	value, err := s.parent.Get(key)
	s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte()*store.Gas(len(key)), GasDescReadPerByte)
	s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte*store.Gas(len(value)), GasDescReadPerByte)

	return value, err
}

func (s *Store) Has(key []byte) (bool, error) {
	s.gasMeter.ConsumeGas(s.gasConfig.HasCost(), GasDescHas)
	return s.parent.Has(key)
}

func (s *Store) Set(key, value []byte) error {
	s.gasMeter.ConsumeGas(s.gasConfig.WriteCostFlat(), GasDescWriteCostFlat)
	s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte()*store.Gas(len(key)), GasDescWritePerByte)
	s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte()*store.Gas(len(value)), GasDescWritePerByte)

	return s.parent.Set(key, value)
}

func (s *Store) Delete(key []byte) error {
	s.gasMeter.ConsumeGas(s.gasConfig.DeleteCost(), GasDescDelete)
	return s.parent.Delete(key)
}

func (b *Store) ApplyChangeSets(changes []store.ChangeSet) error {
	return b.parent.ApplyChangeSets(changes)
}

func (b *Store) ChangeSets() ([]store.ChangeSet, error) {
	return b.parent.ChangeSets()
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	return newIterator(s.parent.Iterator(start, end), s.gasMeter, s.gasConfig)
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	return newIterator(s.parent.ReverseIterator(start, end), s.gasMeter, s.gasConfig)
}
