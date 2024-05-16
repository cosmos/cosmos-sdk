package router

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Service is the interface that wraps the basic methods for a router.
// A router can be a query router or a message router.
type Service interface {
	// CanInvoke returns an error if the given request cannot be invoked.
	CanInvoke(ctx context.Context, typeURL string) error
	// InvokeTyped execute a message or query. It should be used when the called knows the type of the response.
	InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error
	// InvokeUntyped execute a message or query. It should be used when the called doesn't know the type of the response.
	InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error)
}
