package stf

import (
	"context"

	"cosmossdk.io/core/store"
)

var _ store.KVStoreService = (*StoreService)(nil)

func NewStoreService(address []byte) store.KVStoreService {
	return StoreService{address: address}
}

type StoreService struct {
	address []byte
}

func (s StoreService) OpenKVStore(ctx context.Context) store.KVStore {
	state, err := ctx.(*executionContext).store.GetAccountWritableState(s.address)
	if err != nil {
		// tODO: maybe return an erroring store
		panic(err)
	}
	return state
}
