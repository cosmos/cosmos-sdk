package stf

import (
	"context"
	"errors"
)

// getExecutionCtxFromContext tries to get the execution context from the given go context.
func getExecutionCtxFromContext(ctx context.Context) (*executionContext, error) {
	if ec, ok := ctx.(*executionContext); ok {
		return ec, nil
	}

	value, ok := ctx.Value(executionContextKey).(*executionContext)
	if ok {
		return value, nil
	}

	return nil, errors.New("failed to get executionContext from context")
}
