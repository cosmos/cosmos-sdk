package types

import "context"

// CircuitBreaker reports whether a msg type URL may execute. BaseApp checks it
// in the MsgServiceRouter on every msg the router handles. A msg server method
// called directly skips it; msgservice.Dispatch runs a msg through the router
// instead.
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, typeURL string) (bool, error)
}
