package stf

import (
	"context"

	"cosmossdk.io/core/store"
)

var _ store.KVStoreService = (*storeService)(nil)

func NewKVStoreService(address []byte) store.KVStoreService {
	return storeService{actor: address}
}

func NewMemoryStoreService(address []byte) store.MemoryStoreService {
	return storeService{actor: address}
}

type storeService struct {
	actor []byte
}

func (s storeService) OpenKVStore(ctx context.Context) store.KVStore {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		panic(err)
	}

	state, err := exCtx.state.GetWriter(s.actor)
	if err != nil {
		panic(err)
	}
	return state
}

func (s storeService) OpenMemoryStore(ctx context.Context) store.KVStore {
	return s.OpenKVStore(ctx)
}
