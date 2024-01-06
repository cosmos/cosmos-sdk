package mock

import storev2 "cosmossdk.io/store/v2"

var _ storev2.VersionedDatabase = (*StateStorage)(nil)

type StateStorage struct {
}

func (s StateStorage) Has(storeKey string, version uint64, key []byte) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) Get(storeKey string, version uint64, key []byte) ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) GetLatestVersion() (uint64, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) SetLatestVersion(version uint64) error {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) Iterator(storeKey string, version uint64, start, end []byte) (storev2.Iterator, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) ReverseIterator(storeKey string, version uint64, start, end []byte) (storev2.Iterator, error) {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) ApplyChangeset(version uint64, cs *storev2.Changeset) error {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) Prune(version uint64) error {
	// TODO implement me
	panic("implement me")
}

func (s StateStorage) Close() error {
	// TODO implement me
	panic("implement me")
}
