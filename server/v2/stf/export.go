package stf

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
	"cosmossdk.io/core/container"
)

func GetExecutionContext(ctx context.Context) *executionContext {
	_, ok := ctx.(*executionContext)
	if !ok {
		return nil
	}
	return ctx.(*executionContext)
}

func NewExecutionContext() *executionContext {
	executionCtx := &executionContext{Context: context.Background()}
	executionCtx.Cache = NewModuleContainer()
	state := mock.DB()
	executionCtx.State = branch.DefaultNewWriterMap(state)
	return executionCtx
}

func NewStoreService(actor string) store.KVStoreService {
	s := NewKVStoreService([]byte(actor))
	service, ok := s.(interface {
		OpenContainer(ctx context.Context) container.Service
		OpenKVStore(ctx context.Context) store.KVStore
	})
	if ok {
		return service
	}
	return nil
}
