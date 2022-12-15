package runtime

import (
	"context"
	"cosmossdk.io/core/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type kvStoreService struct {
	key *storetypes.KVStoreKey
}

func (k kvStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(k.key)
}

type memStoreService struct {
	key *storetypes.MemoryStoreKey
}

func (m memStoreService) OpenMemoryStore(ctx context.Context) store.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(m.key)
}

type transientStoreService struct {
	key *storetypes.TransientStoreKey
}

func (t transientStoreService) OpenTransientStore(ctx context.Context) store.KVStore {
	return sdk.UnwrapSDKContext(ctx).KVStore(t.key)
}
