package stf

import (
	"context"

	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
)

var _ store.KVStoreService = (*storeService)(nil)

func NewKVStoreService(address []byte) store.KVStoreService {
	return storeService{actor: address}
}

func NewMemoryStoreService(address []byte) store.MemoryStoreService {
	return storeService{actor: address}
}

func NewTransientStoreService(address []byte) store.TransientStoreService {
	return storeService{actor: address}
}

type storeService struct {
	actor []byte
}

func (s storeService) OpenKVStore(ctx context.Context) store.KVStore {
	state, err := ctx.(*executionContext).state.GetWriter(s.actor)
	if err != nil {
		panic(err)
	}
	return state
}

func (s storeService) OpenMemoryStore(ctx context.Context) store.KVStore {
	return s.OpenKVStore(ctx)
}

func (s storeService) OpenTransientStore(ctx context.Context) store.KVStore {
	return s.OpenKVStore(ctx)
}

func NewGasMeterService() gas.Service {
	return gasService{}
}

type gasService struct {
}

func (g gasService) GetGasMeter(ctx context.Context) gas.Meter {
	return ctx.(*executionContext).meter
}

func (g gasService) GetBlockGasMeter(ctx context.Context) gas.Meter {
	panic("stf has no block gas meter")
}

func (g gasService) WithGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	panic("impl")
}

func (g gasService) WithBlockGasMeter(ctx context.Context, meter gas.Meter) context.Context {
	// TODO implement me
	panic("implement me")
}
