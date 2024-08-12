package stf

import (
	"context"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
)

// NewMsgRouterService implements router.Service.
func NewMsgRouterService(identity transaction.Identity) router.Service {
	return msgRouterService{identity: identity}
}

var _ router.Service = (*msgRouterService)(nil)

type msgRouterService struct {
	// TODO(tip): the identity sits here for the purpose of disallowing modules to impersonate others (sudo).
	// right now this is not used, but it serves the reminder of something that we should be eventually
	// looking into.
	identity []byte
}

// CanInvoke returns an error if the given message cannot be invoked.
func (m msgRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	return ctx.(*executionContext).msgRouter.CanInvoke(ctx, typeURL)
}

// Invoke execute a message and returns a response.
func (m msgRouterService) Invoke(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
	return ctx.(*executionContext).msgRouter.Invoke(ctx, msg)
}

// NewQueryRouterService implements router.Service.
func NewQueryRouterService() router.Service {
	return queryRouterService{}
}

var _ router.Service = (*queryRouterService)(nil)

type queryRouterService struct{}

// CanInvoke returns an error if the given request cannot be invoked.
func (m queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	return ctx.(*executionContext).queryRouter.CanInvoke(ctx, typeURL)
}

// Invoke execute a message and returns a response.
func (m queryRouterService) Invoke(
	ctx context.Context,
	req transaction.Msg,
) (transaction.Msg, error) {
	return ctx.(*executionContext).queryRouter.Invoke(ctx, req)
}
