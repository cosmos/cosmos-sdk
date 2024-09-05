package runtime

import (
	"context"

	"cosmossdk.io/core/store"
)

type failingStoreService struct{}

func (failingStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	panic("kv store service not available for this module: verify runtime `skip_store_keys` app config if not expected")
}

func (failingStoreService) OpenMemoryStore(ctx context.Context) store.KVStore {
	panic("memory kv store service not available for this module: verify runtime `skip_store_keys` app config if not expected")
}

func (failingStoreService) OpenTransientStore(ctx context.Context) store.KVStore {
	panic("transient kv store service not available for this module: verify runtime `skip_store_keys` app config if not expected")
}
