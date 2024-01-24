package stf

import (
	"context"

	"cosmossdk.io/core/store"
)

var _ store.KVStoreService = (*StoreService)(nil)

func NewStoreService(address []byte) store.KVStoreService {
	return StoreService{actor: address}
}

type StoreService struct {
	actor []byte
}

func (s StoreService) OpenKVStore(ctx context.Context) store.KVStore {
	state, err := ctx.(*executionContext).store.GetWriter(s.actor)
	if err != nil {
		// tODO: maybe return an erroring store
		panic(err)
	}
	return state
}
