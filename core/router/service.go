package router

import (
	"context"

	"cosmossdk.io/core/transaction"
)

// Service is the interface that wraps the basic methods for a router.
// A router can be a query router or a message router.
type Service interface {
	// CanInvoke returns an error if the given request cannot be invoked.
	CanInvoke(ctx context.Context, typeURL string) error
	// InvokeTyped execute a message or query. It should be used when the called knows the type of the response.
	InvokeTyped(ctx context.Context, req, res transaction.Msg) error
	// InvokeUntyped execute a Msg or query. It should be used when the called doesn't know the type of the response.
	InvokeUntyped(ctx context.Context, req transaction.Msg) (res transaction.Msg, err error)
}
