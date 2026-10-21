package types

import "context"

// CircuitBreaker reports whether a msg type URL may execute. BaseApp checks it
// in the MsgServiceRouter and attaches it to the Context of every tx, block
// hook and query, so msg servers can check msgs that reach them without the
// router.
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, typeURL string) (bool, error)
}
