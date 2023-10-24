package gas

import "cosmossdk.io/store/v2"

var _ store.BranchedKVStore = (*Store)(nil)

type Store struct {
	parent store.KVStore
}

func (s *Store) GetStoreKey() string {
	return s.parent.GetStoreKey()
}

func (s *Store) GetStoreType() store.StoreType {
	return store.StoreTypeGas
}

func (s *Store) Get(key []byte) []byte {
	panic("not implemented!")
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
