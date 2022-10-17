package intermodule

import (
	"context"

	"google.golang.org/grpc"
)

// Client is an inter-module client as specified in ADR-033. It
// allows one module to send msg's and queries to other modules provided
// that the request is valid and can be properly authenticated.
type Client interface {
	grpc.ClientConnInterface

	// InvokerByMethod resolves an invoker for the provided method or returns an error.
	InvokerByMethod(method string) (Invoker, error)

	// InvokerByRequest resolves an invoker for the provided request type or returns an error.
	// This only works for Msg's as they are routed based on type name in transactions already.
	// For queries use InvokerByMethod instead.
	InvokerByRequest(request any) (Invoker, error)

	// DerivedClient returns an inter-module client for the ADR-028 derived
	// module address for the provided key.
	DerivedClient(key []byte) Client

	// Address is the ADR-028 address of this client against which messages will be authenticated.
	Address() []byte
}

// Invoker is an inter-module invoker that has already been resolved to a specific method route.
type Invoker func(ctx context.Context, request any, opts ...grpc.CallOption) (res any, err error)
