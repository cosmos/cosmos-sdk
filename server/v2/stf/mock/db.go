package mock

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

func DB() store.ReadonlyState {
	return memdb{kv: map[string][]byte{}}
}

type memdb struct {
	kv map[string][]byte
}

func (m memdb) Get(key []byte) ([]byte, error) {
	return m.kv[string(key)], nil
}

func (m memdb) Iterator(start, end []byte) (corestore.Iterator, error) {
	panic("not supported")
}

func (m memdb) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("not supported")
}

func (m memdb) Has(key []byte) (bool, error) {
	v, err := m.Get(key)
	return v != nil, err
}
