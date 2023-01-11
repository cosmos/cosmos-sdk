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

type kvStoreWrapper struct {
	sdk.KVStore
}

func (k kvStoreWrapper) Iterator(start, end []byte) store.Iterator {
	return k.KVStore.Iterator(start, end)
}

func (k kvStoreWrapper) ReverseIterator(start, end []byte) store.Iterator {
	return k.KVStore.ReverseIterator(start, end)
}
