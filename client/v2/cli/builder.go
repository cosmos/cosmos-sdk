package cli

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/cosmos/cosmos-sdk/client/v2/cli/flag"
)

type Builder struct {
	flag.Builder
	FileResolver  protodesc.Resolver
	GetClientConn func(context.Context) grpc.ClientConnInterface
}
