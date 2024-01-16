package appmanager

import "context"

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
// TODO: add premessage hook to circuit
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, typeURL string) (bool, error)
}
