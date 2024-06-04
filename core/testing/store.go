package coretesting

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
)

func KVStoreService(ctx context.Context, moduleName string) store.KVStoreService {
	unwrap(ctx).stores[moduleName] = newMemDB()
	return kvStoreService{
		moduleName: moduleName,
	}
}

type kvStoreService struct {
	moduleName string
}

func (k kvStoreService) OpenKVStore(ctx context.Context) store.KVStore {
	kv, ok := unwrap(ctx).stores[k.moduleName]
	if !ok {
		panic(fmt.Sprintf("KVStoreService %s not found", k.moduleName))
	}
	return kv
}
