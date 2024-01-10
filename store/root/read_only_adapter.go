package root

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

var _ store.ReadOnlyRootStore = (*ReadOnlyAdapter)(nil)

type ReadOnlyAdapter struct {
	rootStore store.RootStore
	version   uint64
}

func (roa *ReadOnlyAdapter) Has(storeKey string, key []byte) (bool, error) {
	panic("not implemented yet!")
}

func (roa *ReadOnlyAdapter) Get(storeKey string, key []byte) ([]byte, error) {
	panic("not implemented yet!")
}

func (roa *ReadOnlyAdapter) Iterator(storeKey string, start, end []byte) (corestore.Iterator, error) {
	panic("not implemented yet!")
}

func (roa *ReadOnlyAdapter) ReverseIterator(storeKey string, start, end []byte) (corestore.Iterator, error) {
	panic("not implemented yet!")
}
