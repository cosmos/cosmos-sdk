package testutil

import (
	"context"
	"fmt"

	db "github.com/cosmos/cosmos-db"

	store "cosmossdk.io/collections/corecompat"
)

var _ store.KVStoreService = (*kvStoreService)(nil)

func KVStoreService(ctx context.Context, moduleName string) store.KVStoreService {
	unwrap(ctx).stores[moduleName] = db.NewMemDB()
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
