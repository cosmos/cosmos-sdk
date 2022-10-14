package autocli

import (
	"context"

	"google.golang.org/grpc"

	"cosmossdk.io/client/v2/autocli/flag"
)

// Builder manages options for building CLI commands.
type Builder struct {
	// flag.Builder embeds the flag builder and its options.
	flag.Builder

	// GetClientConn specifies how CLI commands will resolve a grpc.ClientConnInterface
	// from a given context.
	GetClientConn func(context.Context) grpc.ClientConnInterface
}
