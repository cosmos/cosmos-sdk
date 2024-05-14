package stf

import (
	"context"
)

func GetExecutionContext(ctx context.Context) *executionContext {
	executionCtx, ok := ctx.(*executionContext)
	if !ok {
		return nil
	}
	return executionCtx
}
