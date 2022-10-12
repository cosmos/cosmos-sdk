package appmodule

import (
	"context"

	"google.golang.org/grpc"
)

// InterModuleClient is an inter-module client as specified in ADR-033. It
// allows one module to send msg's and queries to other modules provided
// that the request is valid and can be properly authenticated.
type InterModuleClient interface {
	grpc.ClientConnInterface

	// InvokeMsgHandler invokes an arbitrary msg handler given a msg. This is to be used
	// when a module dispatches msg's from another source - such as nested msg's in a transaction.
	// res is expected to be a pointer type and will receive the response.
	InvokeMsgHandler(ctx context.Context, msg interface{}, res interface{}) error

	// Address is the ADR-028 address of this client against which messages will be authenticated.
	Address() []byte
}

// RootInterModuleClient is the root inter-module client of a module which
// uses the ADR-028 root module address.
type RootInterModuleClient interface {
	InterModuleClient

	// DerivedClient returns an inter-module client for the ADR-028 derived
	// module address for the provided key.
	DerivedClient(key []byte) InterModuleClient
}
