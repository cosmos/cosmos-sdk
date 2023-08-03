package counter

import (
	"context"

	counterv1 "cosmossdk.io/api/cosmos/accounts/examples/counter/v1"
	echov1 "cosmossdk.io/api/cosmos/accounts/examples/echo/v1"
	"cosmossdk.io/collections"
	"cosmossdk.io/x/accounts/sdk"
	"google.golang.org/protobuf/proto"
)

func NewCounter(deps *sdk.BuildDependencies) (Counter, error) {
	return Counter{
		Counter: collections.NewSequence(deps.SchemaBuilder, collections.NewPrefix(0), "counter"),
		invoke:  deps.Execute,
	}, nil
}

type Counter struct {
	Counter collections.Sequence

	invoke func(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error)
}

func (a Counter) Init(ctx context.Context, msg *counterv1.MsgInit) (*counterv1.MsgInitResponse, error) {
	err := a.Counter.Set(ctx, msg.CounterValue)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (a Counter) GetCounterValue(ctx context.Context) (uint64, error) {
	return a.Counter.Peek(ctx)
}

func (a Counter) IncreaseCounterValue(ctx context.Context) (uint64, error) {
	return a.Counter.Next(ctx)
}

func (a Counter) Execute(ctx context.Context, target []byte, msg proto.Message) (proto.Message, error) {
	return a.invoke(ctx, target, msg)
}

func (a Counter) RegisterQueryHandlers(router *sdk.QueryRouter) {
	sdk.RegisterQueryHandler(router, func(ctx context.Context, msg *counterv1.QueryCounterRequest) (*counterv1.QueryCounterResponse, error) {
		value, err := a.GetCounterValue(ctx)
		return &counterv1.QueryCounterResponse{CounterValue: value}, err
	})
}

func (a Counter) RegisterExecuteHandlers(router *sdk.ExecuteRouter) {
	sdk.RegisterExecuteHandler(router, func(ctx context.Context, msg *counterv1.MsgIncreaseCounter) (*counterv1.MsgIncreaseCounterResponse, error) {
		newValue, err := a.IncreaseCounterValue(ctx)
		return &counterv1.MsgIncreaseCounterResponse{CounterValue: newValue}, err
	})

	sdk.RegisterExecuteHandler(router, func(ctx context.Context, msg *counterv1.MsgExecuteEcho) (*counterv1.MsgExecuteEchoResponse, error) {
		resp, err := a.invoke(ctx, msg.Target, &echov1.MsgEcho{
			Message: msg.Msg,
		})
		if err != nil {
			return nil, err
		}
		echoResp := resp.(*echov1.MsgEchoResponse)
		return &counterv1.MsgExecuteEchoResponse{
			Result: echoResp.Message,
		}, nil
	})
}

func (a Counter) RegisterInitHandler(router *sdk.InitRouter) {
	sdk.RegisterInitHandler(router, func(ctx context.Context, msg *counterv1.MsgInit) (*counterv1.MsgInitResponse, error) {
		return a.Init(ctx, msg)
	})
}
