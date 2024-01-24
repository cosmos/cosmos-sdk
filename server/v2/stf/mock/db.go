package mock

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
)

func DB() store.ReadonlyAccountsState {
	return accountState{kv: map[string][]byte{}}
}

type accountState struct {
	kv map[string][]byte
}

func (m accountState) GetAccountReadonlyState(address []byte) (store.ReadonlyState, error) {
	return memState{address, m.kv}, nil
}

type memState struct {
	address []byte
	kv      map[string][]byte
}

func (m memState) Has(key []byte) (bool, error) {
	v, err := m.Get(key)
	return v != nil, err
}

func (m memState) Get(bytes []byte) ([]byte, error) {
	key := append(m.address, bytes...)
	return m.kv[string(key)], nil
}

func (m memState) Iterator(start, end []byte) (corestore.Iterator, error) { panic("implement me") }

func (m memState) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	panic("implement me")
}
