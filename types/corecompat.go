package types

import (
	"context"

	"cosmossdk.io/core/store"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

func NewKVStoreService(storeKey *storetypes.KVStoreKey) store.KVStoreService {
	return &kvStoreService{key: storeKey}
}

type kvStoreService struct {
	key *storetypes.KVStoreKey
}

func (k kvStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	return newCoreKVStore(UnwrapSDKContext(ctx).KVStore(k.key))
}

func NewMemStoreService(storeKey *storetypes.MemoryStoreKey) store.MemoryStoreService {
	return &memStoreService{key: storeKey}
}

type memStoreService struct {
	key *storetypes.MemoryStoreKey
}

func (m memStoreService) OpenMemoryStore(ctx context.Context) store.KVStore {
	return newCoreKVStore(UnwrapSDKContext(ctx).KVStore(m.key))
}

func NewTransientStoreService(storeKey *storetypes.TransientStoreKey) store.TransientStoreService {
	return &transientStoreService{key: storeKey}
}

type transientStoreService struct {
	key *storetypes.TransientStoreKey
}

func (t transientStoreService) OpenTransientStore(ctx context.Context) store.KVStore {
	return newCoreKVStore(UnwrapSDKContext(ctx).KVStore(t.key))
}

type coreKVStore struct {
	kvStore storetypes.KVStore
}

func newCoreKVStore(s storetypes.KVStore) store.KVStore {
	return coreKVStore{kvStore: s}
}

func (s coreKVStore) Get(key []byte) ([]byte, error) {
	return s.kvStore.Get(key), nil
}

func (s coreKVStore) Has(key []byte) (bool, error) {
	return s.kvStore.Has(key), nil
}

func (s coreKVStore) Set(key, value []byte) error {
	s.kvStore.Set(key, value)
	return nil
}

func (s coreKVStore) Delete(key []byte) error {
	s.kvStore.Delete(key)
	return nil
}

func (s coreKVStore) Iterator(start, end []byte) (store.Iterator, error) {
	return s.kvStore.Iterator(start, end), nil
}

func (s coreKVStore) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return s.kvStore.ReverseIterator(start, end), nil
}

var _ storetypes.KVStore = kvStoreAdapter{}

type kvStoreAdapter struct {
	store store.KVStore
}

func (kvStoreAdapter) CacheWrap() storetypes.CacheWrap {
	panic("unimplemented")
}

func (kvStoreAdapter) GetStoreType() storetypes.StoreType {
	panic("unimplemented")
}

func (s kvStoreAdapter) Delete(key []byte) {
	if err := s.store.Delete(key); err != nil {
		panic(err)
	}
}

func (s kvStoreAdapter) Get(key []byte) []byte {
	bz, err := s.store.Get(key)
	if err != nil {
		panic(err)
	}
	return bz
}

func (s kvStoreAdapter) Has(key []byte) bool {
	has, err := s.store.Has(key)
	if err != nil {
		panic(err)
	}
	return has
}

func (s kvStoreAdapter) Set(key, value []byte) {
	if err := s.store.Set(key, value); err != nil {
		panic(err)
	}
}

func (s kvStoreAdapter) Iterator(start, end []byte) storetypes.Iterator {
	it, err := s.store.Iterator(start, end)
	if err != nil {
		panic(err)
	}
	return it
}

func (s kvStoreAdapter) ReverseIterator(start, end []byte) storetypes.Iterator {
	it, err := s.store.ReverseIterator(start, end)
	if err != nil {
		panic(err)
	}
	return it
}

func KVStoreAdapter(store store.KVStore) storetypes.KVStore {
	return &kvStoreAdapter{store}
}
