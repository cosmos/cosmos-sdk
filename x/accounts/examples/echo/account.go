package echo

import (
	"context"

	v1 "cosmossdk.io/x/accounts/examples/echo/v1"
	"cosmossdk.io/x/accounts/sdk"
)

type Echo struct{}

func NewEcho(_ *sdk.BuildDependencies) (Echo, error) {
	return Echo{}, nil
}

func (a Echo) RegisterInitHandler(_ *sdk.InitRouter) {}
func (a Echo) RegisterExecuteHandlers(r *sdk.ExecuteRouter) {
	sdk.RegisterExecuteHandler(r, func(ctx context.Context, msg v1.MsgEcho) (v1.MsgEchoResponse, error) {
		return v1.MsgEchoResponse{
			Message: msg.Message,
			Sender:  sdk.Sender(ctx),
		}, nil
	})
}

func (a Echo) RegisterQueryHandlers(r *sdk.QueryRouter) {
	sdk.RegisterQueryHandler(r, func(ctx context.Context, msg v1.QueryEchoRequest) (v1.QueryEchoResponse, error) {
		return v1.QueryEchoResponse{
			Message: msg.Message,
		}, nil
	})
}
