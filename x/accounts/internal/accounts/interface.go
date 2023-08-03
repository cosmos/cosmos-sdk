package accounts

import (
	"context"

	"cosmossdk.io/collections"
	"google.golang.org/protobuf/proto"
)

// Account defines a generic account interface.
type Account interface {
	// RegisterInitHandler must be used by the account to register its initialisation handler.
	RegisterInitHandler(router *InitRouter)
	// RegisterExecuteHandlers is given a router and the account should register
	// its execute handlers.
	RegisterExecuteHandlers(router *ExecuteRouter)
	// RegisterQueryHandlers is given a router and the account should register
	// its query handlers.
	RegisterQueryHandlers(router *QueryRouter)
}

type BuildDependencies struct {
	SchemaBuilder *collections.SchemaBuilder
	Execute       func(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error)
	Query         func(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error)
}

type Msg[T any] interface {
	*T
	proto.Message
}
