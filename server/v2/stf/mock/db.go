package mock

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

func DB() store.ReadonlyState {
	return memState{kv: map[string][]byte{}}
}

type memState struct {
	kv map[string][]byte
}

func (m memState) Get(key []byte) ([]byte, error) {
	return m.kv[string(key)], nil
}

func (m memState) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not supported")
}

func (m memState) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not supported")
}

func (m memState) Has(key []byte) (bool, error) {
	v, err := m.Get(key)
	return v != nil, err
}
