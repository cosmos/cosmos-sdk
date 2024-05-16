package stf

import (
	"context"

	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/server/v2/stf/mock"
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
	executionCtx.state = branch.DefaultNewWriterMap(state)
	return executionCtx
}

