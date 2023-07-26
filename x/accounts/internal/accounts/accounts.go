package accounts

import (
	"context"
	"fmt"

	"cosmossdk.io/x/accounts/sdk"
	"github.com/cosmos/gogoproto/proto"
)

func NewAccountImplementation[
	IReq, IResp any, IReqP sdk.Msg[IReq], IRespP sdk.Msg[IResp],
](typeName string, genericAccount sdk.Account[IReq, IResp, IReqP, IRespP]) (Implementation, error) {
	executeRouter := &sdk.ExecuteRouter{}
	err := genericAccount.RegisterExecuteHandlers(executeRouter)
	if err != nil {
		return Implementation{}, err
	}

	queryRouter := &sdk.QueryRouter{}
	err = genericAccount.RegisterQueryHandlers(queryRouter)
	if err != nil {
		return Implementation{}, err
	}
	return Implementation{
		Execute: executeRouter.Handler(),
		Query:   queryRouter.Handler(),
		Init: func(ctx context.Context, msg proto.Message) (proto.Message, error) {
			methodName := proto.MessageName(IRespP(new(IResp)))
			concrete, ok := msg.(IReqP)
			if !ok {
				return nil, fmt.Errorf("invalid message type %T, wanted: %s", msg, methodName)
			}
			resp, err := genericAccount.Init(ctx, *concrete)
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
	Execute func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Query   func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Init    func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Type    func() string
}
