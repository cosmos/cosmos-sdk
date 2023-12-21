package appmanager

import "context"

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
// TODO replace with pre message hooks
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, typeURL string) (bool, error)
}
