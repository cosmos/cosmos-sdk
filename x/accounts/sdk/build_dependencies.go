package sdk

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/gogoproto/proto"
)

type BuildDependencies struct {
	SchemaBuilder *collections.SchemaBuilder
	Invoker       func(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error)
}
