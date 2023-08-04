package accounts

import (
	"context"
	"fmt"

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

func Invoke[Resp, Req any, RespP Msg[Resp], ReqP Msg[Req]](client Invoker, ctx context.Context, target []byte, req ReqP) (RespP, error) {
	resp, err := client.invoke(ctx, target, req)
	if err != nil {
		return nil, err
	}
	concreteResp, ok := resp.(RespP)
	if !ok {
		return nil, fmt.Errorf(
			"unexpected response type: %s, wanted: %s",
			proto.MessageName(resp),
			proto.MessageName(RespP(new(Resp))),
		)
	}
	return concreteResp, nil
}

type Invoker struct {
	invoke func(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error)
}

type BuildDependencies struct {
	SchemaBuilder *collections.SchemaBuilder
	Execute       Invoker
	Query         func(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error)
}

type Msg[T any] interface {
	*T
	proto.Message
}
