package router

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Router embeds a QueryRouterService and MessageRouterService.
// Each service allows to invoke messages and queries via the corresponding router.
type Router interface {
	QueryRouterService() Service
	MessageRouterService() Service
}

// Service is the interface that wraps the basic methods for a router service.
type Service interface {
	// InvokeTyped execute a message or query. It should be used when the called knows the type of the response.
	InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error
	// InvokeUntyped execute a message or query. It should be used when the called doesn't know the type of the response.
	InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error)
}
