package root

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

var (
	_ corestore.Reader    = (*Reader)(nil)
	_ corestore.ReaderMap = (*ReaderMap)(nil)
)

// ReaderMap defines an adapter around a RootStore that only exposes read-only
// operations. This is useful for exposing a read-only view of the RootStore at
// a specific version in history, which could also be the latest state.
type ReaderMap struct {
	rootStore store.RootStore
	version   uint64
}

func NewReaderMap(v uint64, rs store.RootStore) *ReaderMap {
	return &ReaderMap{
		rootStore: rs,
		version:   v,
	}
}

func (roa *ReaderMap) GetReader(actor []byte) (corestore.Reader, error) {
	return NewReader(roa.version, roa.rootStore, actor), nil
}

// Reader represents a read-only adapter for accessing data from the root store.
type Reader struct {
	version   uint64          // The version of the data.
	rootStore store.RootStore // The root store to read data from.
	actor     []byte          // The actor associated with the data.
}

func NewReader(v uint64, rs store.RootStore, actor []byte) *Reader {
	return &Reader{
		version:   v,
		rootStore: rs,
		actor:     actor,
	}
}

func (roa *Reader) Has(key []byte) (bool, error) {
	val, err := roa.rootStore.GetStateStorage().Has(roa.actor, roa.version, key)
	if err != nil {
		return false, err
	}

	return val, nil
}

func (roa *Reader) Get(key []byte) ([]byte, error) {
	result, err := roa.rootStore.GetStateStorage().Get(roa.actor, roa.version, key)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (roa *Reader) Iterator(start, end []byte) (corestore.Iterator, error) {
	return roa.rootStore.GetStateStorage().Iterator(roa.actor, roa.version, start, end)
}

func (roa *Reader) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return roa.rootStore.GetStateStorage().ReverseIterator(roa.actor, roa.version, start, end)
}
