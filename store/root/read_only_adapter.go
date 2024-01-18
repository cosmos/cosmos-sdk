package root

import (
	corestore "cosmossdk.io/core/store"
	corestorev2 "cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/store/v2"
)

var _ corestorev2.ReadonlyState = (*ReadOnlyAdapter)(nil)

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

func (roa *ReadOnlyAdapter) Has(key []byte) (bool, error) {
	val, err := roa.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (roa *ReadOnlyAdapter) Get(key []byte) ([]byte, error) {
	storeKey := "main" // TODO extract store key, check if fine with store team
	result, err := roa.rootStore.Query(storeKey, roa.version, key, false)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

func (roa *ReadOnlyAdapter) Iterator(start, end []byte) (corestore.Iterator, error) {
	storeKey := "main" // TODO extract store key
	return roa.rootStore.GetStateStorage().Iterator(storeKey, roa.version, start, end)
}

func (roa *ReadOnlyAdapter) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	storeKey := "main" // TODO extract store key
	return roa.rootStore.GetStateStorage().ReverseIterator(storeKey, roa.version, start, end)
}
