package module

import (
	"context"

	"cosmossdk.io/core/store"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewCoreKVStoreService(key *sdk.KVStoreKey) store.KVStoreService {
	return kvStoreService{key}
}

type kvStoreService struct {
	key *sdk.KVStoreKey
}

func (k kvStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	return kvStoreWrapper{sdk.UnwrapSDKContext(ctx).KVStore(k.key)}
}

func NewCoreMemStoreService(key *sdk.MemoryStoreKey) store.MemoryStoreService {
	return memStoreService{key}
}

type memStoreService struct {
	key *sdk.MemoryStoreKey
}

func (k memStoreService) OpenMemoryStore(ctx context.Context) store.KVStore {
	return kvStoreWrapper{sdk.UnwrapSDKContext(ctx).KVStore(k.key)}
}

func NewCoreTransientStoreService(key *sdk.TransientStoreKey) store.TransientStoreService {
	return transientStoreService{}
}

type transientStoreService struct {
	key *sdk.TransientStoreKey
}

func (k transientStoreService) OpenTransientStore(ctx context.Context) store.KVStore {
	return kvStoreWrapper{sdk.UnwrapSDKContext(ctx).KVStore(k.key)}
}

var _ store.KVStore = kvStoreWrapper{}

type kvStoreWrapper struct {
	sdk.KVStore
}

func (k kvStoreWrapper) Iterator(start, end []byte) (store.Iterator, error) {
	return k.KVStore.Iterator(start, end), nil
}

func (k kvStoreWrapper) ReverseIterator(start, end []byte) (store.Iterator, error) {
	return k.KVStore.ReverseIterator(start, end), nil
}

// Get returns nil iff key doesn't exist. Errors on nil key.
func (k kvStoreWrapper) Get(key []byte) ([]byte, error) {
	return k.KVStore.Get(key), nil
}

// Has checks if a key exists. Errors on nil key.
func (k kvStoreWrapper) Has(key []byte) (bool, error) {
	return k.KVStore.Has(key), nil
}

// Set sets the key. Errors on nil key or value.
func (k kvStoreWrapper) Set(key, value []byte) error {
	k.KVStore.Set(key, value)
	return nil
}

// Delete deletes the key. Errors on nil key.
func (k kvStoreWrapper) Delete(key []byte) error {
	k.KVStore.Delete(key)
	return nil
}
