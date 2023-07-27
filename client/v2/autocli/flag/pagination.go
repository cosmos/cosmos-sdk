package flag

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/client/v2/internal/util"
)

func (b *Builder) bindPageRequest(ctx context.Context, flagSet *pflag.FlagSet, field protoreflect.FieldDescriptor) (HasValue, error) {
	return b.addMessageFlags(
		ctx,
		flagSet,
		util.ResolveMessageType(b.TypeResolver, field.Message()),
		&autocliv1.RpcCommandOptions{},
		namingOptions{Prefix: "page-"},
	)
}
