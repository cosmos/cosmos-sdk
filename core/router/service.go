package router

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"
)

// Service is the interface that wraps the basic methods for a router.
// A router can be a query router or a message router.
type Service interface {
	// CanInvoke returns an error if the given request cannot be invoked.
	CanInvoke(ctx context.Context, typeURL string) error
	// InvokeTyped execute a message or query. It should be used when the called knows the type of the response.
	InvokeTyped(ctx context.Context, req, res gogoproto.Message) error
	// InvokeUntyped execute a message or query. It should be used when the called doesn't know the type of the response.
	InvokeUntyped(ctx context.Context, req gogoproto.Message) (res gogoproto.Message, err error)
}
