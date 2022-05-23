package cli

import (
	"context"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/client/v2/cli/flag"
)

// Builder manages options for building CLI commands.
type Builder struct {
	// flag.Builder embeds the flag builder and its options.
	flag.Builder

	// GetClientConn specifies how CLI commands will resolve a grpc.ClientConnInterface
	// from a given context.
	GetClientConn func(context.Context) grpc.ClientConnInterface
}
