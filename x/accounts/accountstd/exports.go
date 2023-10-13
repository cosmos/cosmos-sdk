package accountstd

import (
	"context"

	"cosmossdk.io/x/accounts/internal/implementation"
)

// Interface is the exported interface of an Account.
type Interface = implementation.Account

// ExecuteBuilder is the exported type of ExecuteBuilder.
type ExecuteBuilder = implementation.ExecuteBuilder

// QueryBuilder is the exported type of QueryBuilder.
type QueryBuilder = implementation.QueryBuilder

// InitBuilder is the exported type of InitBuilder.
type InitBuilder = implementation.InitBuilder

// AccountCreatorFunc is the exported type of AccountCreatorFunc.
type AccountCreatorFunc = implementation.AccountCreatorFunc

// Dependencies is the exported type of Dependencies.
type Dependencies = implementation.Dependencies

func RegisterExecuteHandler[
	Req any, ProtoReq implementation.ProtoMsg[Req], Resp any, ProtoResp implementation.ProtoMsg[Resp],
](router *ExecuteBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	implementation.RegisterExecuteHandler(router, handler)
}

// RegisterQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterQueryHandler[
	Req any, ProtoReq implementation.ProtoMsg[Req], Resp any, ProtoResp implementation.ProtoMsg[Resp],
](router *QueryBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	implementation.RegisterQueryHandler(router, handler)
}

// RegisterInitHandler registers an initialisation handler for a smart account that uses protobuf.
func RegisterInitHandler[
	Req any, ProtoReq implementation.ProtoMsg[Req], Resp any, ProtoResp implementation.ProtoMsg[Resp],
](router *InitBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	implementation.RegisterInitHandler(router, handler)
}

// AddAccount is a helper function to add a smart account to the list of smart accounts.
func AddAccount[A Interface](name string, constructor func(deps Dependencies) (A, error)) AccountCreatorFunc {
	return implementation.AddAccount(name, constructor)
}
