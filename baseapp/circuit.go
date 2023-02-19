package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
type CircuitBreaker interface {
	IsAllowed(ctx sdk.Context, typeURL string) bool
}
