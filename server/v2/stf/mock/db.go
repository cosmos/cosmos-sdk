package mock

import (
	"cosmossdk.io/core/store"
)

func DB() store.ReaderMap {
	return actorState{kv: map[string][]byte{}}
}

type actorState struct {
	kv map[string][]byte
}

func (m actorState) GetReader(address []byte) (store.Reader, error) {
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

func (m memState) Iterator(start, end []byte) (store.Iterator, error) { panic("implement me") }

func (m memState) ReverseIterator(start, end []byte) (store.Iterator, error) {
	panic("implement me")
}
