package root

import (
	corestore "cosmossdk.io/core/store"
	corestorev2 "cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/store/v2"
)

var _ corestorev2.Reader = (*Reader)(nil)
var _ corestorev2.ReaderMap = (*ReaderMap)(nil)

// ReaderMap defines an adapter around a RootStore that only exposes read-only
// operations. This is useful for exposing a read-only view of the RootStore at
// a specific version in history, which could also be the latest state.
type ReaderMap struct {
	rootStore store.RootStore
	actor     []byte
	version   uint64
}

func NewReadOnlyAdapter(v uint64, rs store.RootStore) *ReaderMap {
	return &ReaderMap{
		rootStore: rs,
		version:   v,
	}
}

func (roa *ReaderMap) GetReader(actor []byte) (Reader, error) {
	return *NewReader(roa.version, roa.rootStore, actor), nil
}

type Reader struct {
	version   uint64
	rootStore store.RootStore
	actor     []byte
}

func NewReader(v uint64, rs store.RootStore, actor []byte) *Reader {
	return &Reader{
		version:   v,
		rootStore: rs,
		actor:     actor,
	}
}

func (roa *Reader) Has(key []byte) (bool, error) {
	val, err := roa.Has(key)
	if err != nil {
		return false, err
	}

	return val, nil
}

func (roa *Reader) Get(key []byte) ([]byte, error) {
	result, err := roa.Get(key)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (roa *Reader) Iterator(start, end []byte) (corestore.Iterator, error) {
	return roa.rootStore.GetStateStorage().Iterator(string(roa.actor), roa.version, start, end)
}

func (roa *Reader) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return roa.rootStore.GetStateStorage().ReverseIterator(string(roa.actor), roa.version, start, end)
}
