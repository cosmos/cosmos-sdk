package root

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

var _ store.ReadOnlyRootStore = (*ReadOnlyAdapter)(nil)

// ReadOnlyAdapter defines an adapter around a RootStore that only exposes read-only
// operations. This is useful for exposing a read-only view of the RootStore at
// a specific version in history, which could also be the latest state.
type ReadOnlyAdapter struct {
	rootStore store.RootStore
	version   uint64
}

func NewReadOnlyAdapter(v uint64, rs store.RootStore) *ReadOnlyAdapter {
	return &ReadOnlyAdapter{
		rootStore: rs,
		version:   v,
	}
}

func (roa *ReadOnlyAdapter) Has(storeKey string, key []byte) (bool, error) {
	val, err := roa.Get(storeKey, key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (roa *ReadOnlyAdapter) Get(storeKey string, key []byte) ([]byte, error) {
	result, err := roa.rootStore.Query(storeKey, roa.version, key, false)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

func (roa *ReadOnlyAdapter) Iterator(storeKey string, start, end []byte) (corestore.Iterator, error) {
	return roa.rootStore.GetStateStorage().Iterator(storeKey, roa.version, start, end)
}

func (roa *ReadOnlyAdapter) ReverseIterator(storeKey string, start, end []byte) (corestore.Iterator, error) {
	return roa.rootStore.GetStateStorage().ReverseIterator(storeKey, roa.version, start, end)
}
