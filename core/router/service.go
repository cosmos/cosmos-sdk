package router

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Service is the interface that wraps the basic methods for a router service.
type Service interface {
	InvokeTyped(ctx context.Context, req, res protoiface.MessageV1) error
	// InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (res protoiface.MessageV1, err error)
}
