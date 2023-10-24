package gas

import "cosmossdk.io/store/v2"

var _ store.BranchedKVStore = (*Store)(nil)

type Store struct {
	parent    store.KVStore
	gasMeter  store.GasMeter
	gasConfig store.GasConfig
}

func New(p store.KVStore, gm store.GasMeter, gc store.GasConfig) store.BranchedKVStore {
	return &Store{
		parent:    p,
		gasMeter:  gm,
		gasConfig: gc,
	}
}

func (s *Store) GetStoreKey() string {
	return s.parent.GetStoreKey()
}

func (s *Store) GetStoreType() store.StoreType {
	return s.parent.GetStoreType()
}

func (s *Store) Get(key []byte) []byte {
	s.gasMeter.ConsumeGas(s.gasConfig.ReadCostFlat, store.GasDescReadCostFlat)

	value := s.parent.Get(key)
	s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte*store.Gas(len(key)), store.GasDescReadPerByte)
	s.gasMeter.ConsumeGas(s.gasConfig.ReadCostPerByte*store.Gas(len(value)), store.GasDescReadPerByte)

	return value
}

func (s *Store) Has(key []byte) bool {
	s.gasMeter.ConsumeGas(s.gasConfig.HasCost, store.GasDescHas)
	return s.parent.Has(key)
}

func (s *Store) Set(key, value []byte) {
	s.gasMeter.ConsumeGas(s.gasConfig.WriteCostFlat, store.GasDescWriteCostFlat)
	s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte*store.Gas(len(key)), store.GasDescWritePerByte)
	s.gasMeter.ConsumeGas(s.gasConfig.WriteCostPerByte*store.Gas(len(value)), store.GasDescWritePerByte)
	s.parent.Set(key, value)
}

func (s *Store) Delete(key []byte) {
	s.gasMeter.ConsumeGas(s.gasConfig.DeleteCost, store.GasDescDelete)
	s.parent.Delete(key)
}

func (s *Store) GetChangeset() *store.Changeset {
	return s.parent.GetChangeset()
}

func (s *Store) Reset() error {
	return s.parent.Reset()
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}
