package stf

import (
	"context"

	"cosmossdk.io/core/store"
)

func GetExecutionContext(ctx context.Context) *executionContext {
	executionCtx, ok := ctx.(*executionContext)
	if !ok {
		return nil
	}
	return executionCtx
}

func GetStateFromContext(ctx *executionContext) store.WriterMap {
	return ctx.state
}
