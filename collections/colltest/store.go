package colltest

import (
	"context"

	db "github.com/cosmos/cosmos-db"

	"cosmossdk.io/core/store"
)

type contextStoreKey struct{}

// MockStore returns a mock store.KVStoreService and a mock context.Context.
// They can be used to test collections. The StoreService.NewStoreContext
// can be used to instantiate a new empty KVStore.
func MockStore() (*StoreService, context.Context) {
	kv := db.NewMemDB()
	ctx := context.WithValue(context.Background(), contextStoreKey{}, &testStore{kv})
	return &StoreService{}, ctx
}

type StoreService struct{}

func (s StoreService) OpenKVStore(ctx context.Context) store.KVStore {
	return ctx.Value(contextStoreKey{}).(store.KVStore)
}

func (s StoreService) NewStoreContext() context.Context {
	kv := db.NewMemDB()
	return context.WithValue(context.Background(), contextStoreKey{}, &testStore{kv})
}

type testStore struct {
	db db.DB
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
