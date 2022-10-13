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

	InvokerByMethod(method string) (Invoker, error)
	InvokerByRequest(request any) (Invoker, error)

	// DerivedClient returns an inter-module client for the ADR-028 derived
	// module address for the provided key.
	DerivedClient(key []byte) Client

	// Address is the ADR-028 address of this client against which messages will be authenticated.
	Address() []byte
}

type Invoker func(ctx context.Context, request any, opts ...grpc.CallOption) (res any, err error)
