package stf

import (
	"context"
	"sync"

	"cosmossdk.io/core/store"
)

var _ store.KVStoreService = (*storeService)(nil)

func NewKVStoreService(address []byte) store.KVStoreService {
	return storeService{actor: address, mu: &sync.RWMutex{}}
}

func NewMemoryStoreService(address []byte) store.MemoryStoreService {
	return storeService{actor: address, mu: &sync.RWMutex{}}
}

type storeService struct {
	actor []byte
	mu    *sync.RWMutex
}

func (s storeService) OpenKVStore(ctx context.Context) store.KVStore {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		panic(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := exCtx.state.GetWriter(s.actor)
	if err != nil {
		panic(err)
	}

	return state
}

func (s storeService) OpenMemoryStore(ctx context.Context) store.KVStore {
	return s.OpenKVStore(ctx)
}
