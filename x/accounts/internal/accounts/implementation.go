package accounts

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/sdk"
	"github.com/cosmos/gogoproto/proto"
)

type Msg[T any] interface {
	*T
	proto.Message
}

func NewAccountImplementation[
	IReq, IResp any, IReqP Msg[IReq], IRespP Msg[IResp],
	Account sdk.Account[IReq, IResp, IReqP, IRespP],
](
	typeName string,
	deps *sdk.BuildDependencies,
	constructor func(sb *sdk.BuildDependencies) (Account, error),
) (Implementation, error) {
	account, err := constructor(deps)
	if err != nil {
		return Implementation{}, err
	}
	executeRouter := &sdk.ExecuteRouter{}
	err = account.RegisterExecuteHandlers(executeRouter)
	if err != nil {
		return Implementation{}, err
	}

	queryRouter := &sdk.QueryRouter{}
	err = account.RegisterQueryHandlers(queryRouter)
	if err != nil {
		return Implementation{}, err
	}

	// build schema
	schema, err := deps.SchemaBuilder.Build()
	if err != nil {
		return Implementation{}, err
	}
	return Implementation{
		StateSchema: schema,
		Execute:     executeRouter.Handler(),
		Query:       queryRouter.Handler(),
		Init: func(ctx context.Context, msg proto.Message) (proto.Message, error) {
			methodName := proto.MessageName(IRespP(new(IResp)))
			concrete, ok := msg.(IReqP)
			if !ok {
				return nil, fmt.Errorf("invalid message type %T, wanted: %s", msg, methodName)
			}
			resp, err := account.Init(ctx, *concrete)
			if err != nil {
				return nil, err
			}
			return IRespP(&resp), nil
		},
		Type: func() string {
			return typeName
		},
	}, nil

}

// Implementation represents the implementation of the Accounts module.
type Implementation struct {
	StateSchema collections.Schema
	Execute     func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Query       func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Init        func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Type        func() string
}
