package colltest

import (
	"context"

	"cosmossdk.io/core/store"
	db "github.com/cosmos/cosmos-db"
)

// MockStore returns a mock store.KVStoreService and a mock context.Context.
// They can be used to test collections.
func MockStore() (store.KVStoreService, context.Context) {
	kv := db.NewMemDB()
	return &testStore{kv}, context.Background()
}

type testStore struct {
	db db.DB
}

func (t testStore) OpenKVStore(ctx context.Context) store.KVStore {
	return t
}

func (t testStore) Get(key []byte) ([]byte, error) {
	return t.db.Get(key)
}

func (t testStore) Has(key []byte) (bool, error) {
	return t.db.Has(key)
}

func (t testStore) Set(key, value []byte) error {
	return t.db.Set(key, value)
}

func (t testStore) Delete(key []byte) error {
	return t.db.Delete(key)
}

func (t testStore) Iterator(start, end []byte) (store.Iterator, error) {
	return t.db.Iterator(start, end)
}

func (t testStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return t.db.ReverseIterator(start, end)
}

var _ store.KVStore = testStore{}
