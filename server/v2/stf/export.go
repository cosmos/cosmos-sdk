package stf

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
)

func GetExecutionContext(ctx context.Context) *executionContext {
	executionCtx, ok := ctx.(*executionContext)
	if !ok {
		return nil
	}
	return executionCtx
}

func NewExecutionContext() *executionContext {
	executionCtx := &executionContext{}
	executionCtx.Cache = NewModuleContainer()
	state := mock.DB()
	executionCtx.State = branch.DefaultNewWriterMap(state)
	return executionCtx
}

func NewStoreService(actor string) store.KVStoreService {
	s := NewKVStoreService([]byte(actor))
	service, ok := s.(interface {
		OpenContainer(ctx context.Context) Container
		OpenKVStore(ctx context.Context) store.KVStore
	})
	if ok {
		return service
	}
	return nil
}
