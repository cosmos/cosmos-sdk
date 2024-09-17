package mock

import (
	corestore "cosmossdk.io/core/store"
)

// ReaderMap defines an adapter around a RootStore that only exposes read-only
// operations. This is useful for exposing a read-only view of the RootStore at
// a specific version in history, which could also be the latest state.
type ReaderMap struct {
	store   *MockStore
	version uint64
}

func NewMockReaderMap(v uint64, rs *MockStore) *ReaderMap {
	return &ReaderMap{
		store:   rs,
		version: v,
	}
}

func (roa *ReaderMap) GetReader(actor []byte) (corestore.Reader, error) {
	return NewMockReader(roa.version, roa.store, actor), nil
}

// Reader represents a read-only adapter for accessing data from the root store.
type MockReader struct {
	version uint64     // The version of the data.
	store   *MockStore // The root store to read data from.
	actor   []byte     // The actor associated with the data.
}

func NewMockReader(v uint64, rs *MockStore, actor []byte) *MockReader {
	return &MockReader{
		version: v,
		store:   rs,
		actor:   actor,
	}
}

func (roa *MockReader) Has(key []byte) (bool, error) {
	val, err := roa.store.GetStateStorage().Has(roa.actor, roa.version, key)
	if err != nil {
		return false, err
	}

	return val, nil
}

func (roa *MockReader) Get(key []byte) ([]byte, error) {
	result, err := roa.store.GetStateStorage().Get(roa.actor, roa.version, key)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (roa *MockReader) Iterator(start, end []byte) (corestore.Iterator, error) {
	return roa.store.GetStateStorage().Iterator(roa.actor, roa.version, start, end)
}

func (roa *MockReader) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return roa.store.GetStateStorage().ReverseIterator(roa.actor, roa.version, start, end)
}
