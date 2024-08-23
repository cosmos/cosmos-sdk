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
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return err
	}

	return exCtx.msgRouter.CanInvoke(ctx, typeURL)
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m msgRouterService) InvokeTyped(ctx context.Context, msg, resp transaction.Msg) error {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return err
	}

	return exCtx.msgRouter.InvokeTyped(ctx, msg, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m msgRouterService) InvokeUntyped(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return exCtx.msgRouter.InvokeUntyped(ctx, msg)
}

// NewQueryRouterService implements router.Service.
func NewQueryRouterService() router.Service {
	return queryRouterService{}
}

var _ router.Service = (*queryRouterService)(nil)

type queryRouterService struct{}

// CanInvoke returns an error if the given request cannot be invoked.
func (m queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return err
	}

	return exCtx.queryRouter.CanInvoke(ctx, typeURL)
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m queryRouterService) InvokeTyped(
	ctx context.Context,
	req, resp transaction.Msg,
) error {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return err
	}

	return exCtx.queryRouter.InvokeTyped(ctx, req, resp)
}

// InvokeUntyped execute a message and returns a response.
func (m queryRouterService) InvokeUntyped(
	ctx context.Context,
	req transaction.Msg,
) (transaction.Msg, error) {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return exCtx.queryRouter.InvokeUntyped(ctx, req)
}
