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
	panic("not implemented!")
}

func (s *Store) Set(key, value []byte) {
	panic("not implemented!")
}

func (s *Store) Delete(key []byte) {
	panic("not implemented!")
}

func (s *Store) GetChangeset() *store.Changeset {
	panic("not implemented!")
}

func (s *Store) Reset() error {
	panic("not implemented!")
}

func (s *Store) Iterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}

func (s *Store) ReverseIterator(start, end []byte) store.Iterator {
	panic("not implemented!")
}
