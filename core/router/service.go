package router

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Service embeds a QueryRouterService and MessageRouterService.
// Each router allows to invoke messages and queries via the corresponding router.
type Service interface {
	QueryRouterService() Router
	MessageRouterService() Router
}

// Router is the interface that wraps the basic methods for a router.
type Router interface {
	// CanInvoke returns an error if the given request cannot be invoked.
	CanInvoke(ctx context.Context, typeURL string) error
	// InvokeTyped execute a message or query. It should be used when the called knows the type of the response.
	InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error
	// InvokeUntyped execute a message or query. It should be used when the called doesn't know the type of the response.
	InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error)
}
