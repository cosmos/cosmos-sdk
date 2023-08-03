package echo

import (
	"context"

	echov1 "cosmossdk.io/api/cosmos/accounts/examples/echo/v1"
	"cosmossdk.io/x/accounts/sdk"
)

type Echo struct{}

func NewEcho(_ *sdk.BuildDependencies) (Echo, error) {
	return Echo{}, nil
}

func (a Echo) RegisterInitHandler(_ *sdk.InitRouter) {}
func (a Echo) RegisterExecuteHandlers(r *sdk.ExecuteRouter) {
	sdk.RegisterExecuteHandler(r, func(ctx context.Context, msg *echov1.MsgEcho) (*echov1.MsgEchoResponse, error) {
		return &echov1.MsgEchoResponse{
			Message: msg.Message,
			Sender:  sdk.Sender(ctx),
		}, nil
	})
}

func (a Echo) RegisterQueryHandlers(r *sdk.QueryRouter) {
	sdk.RegisterQueryHandler(r, func(ctx context.Context, msg *echov1.QueryEchoRequest) (*echov1.QueryEchoResponse, error) {
		return &echov1.QueryEchoResponse{
			Message: msg.Message,
		}, nil
	})
}
