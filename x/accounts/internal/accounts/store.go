package accounts

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/prefixstore"
)

type prefixedStorage struct{}

func SetPrefixedStorage(ctx context.Context, prefix []byte) {
	context.WithValue(ctx, prefixedStorage{}, prefix)
}

func PrefixedStorageService(svc store.KVStoreService) store.KVStoreService {
	return prefixedKVStoreSvc{
		svc: svc,
	}
}

type prefixedKVStoreSvc struct {
	svc store.KVStoreService
}

func (s prefixedKVStoreSvc) OpenKVStore(ctx context.Context) store.KVStore {
	storePrefix := ctx.Value(prefixedStorage{}).([]byte)
	return prefixstore.NewStore(s.svc.OpenKVStore(ctx), storePrefix)
}
