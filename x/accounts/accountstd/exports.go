// Package accountstd exports the types and functions that are used by developers to implement smart accounts.
package accountstd

import (
	"bytes"
	"context"
	"fmt"

	"cosmossdk.io/x/accounts/internal/implementation"

	"github.com/cosmos/cosmos-sdk/types/address"
)

var accountsModuleAddress = address.Module("accounts")

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
	Req any, ProtoReq implementation.ProtoMsgG[Req], Resp any, ProtoResp implementation.ProtoMsgG[Resp],
](router *ExecuteBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	implementation.RegisterExecuteHandler(router, handler)
}

// RegisterQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterQueryHandler[
	Req any, ProtoReq implementation.ProtoMsgG[Req], Resp any, ProtoResp implementation.ProtoMsgG[Resp],
](router *QueryBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	implementation.RegisterQueryHandler(router, handler)
}

// RegisterInitHandler registers an initialisation handler for a smart account that uses protobuf.
func RegisterInitHandler[
	Req any, ProtoReq implementation.ProtoMsgG[Req], Resp any, ProtoResp implementation.ProtoMsgG[Resp],
](router *InitBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	implementation.RegisterInitHandler(router, handler)
}

// AddAccount is a helper function to add a smart account to the list of smart accounts.
func AddAccount[A Interface](name string, constructor func(deps Dependencies) (A, error)) AccountCreatorFunc {
	return func(deps implementation.Dependencies) (string, implementation.Account, error) {
		acc, err := constructor(deps)
		return name, acc, err
	}
}

// Whoami returns the address of the account being invoked.
func Whoami(ctx context.Context) []byte {
	return implementation.Whoami(ctx)
}

// Sender returns the sender of the execution request.
func Sender(ctx context.Context) []byte {
	return implementation.Sender(ctx)
}

// SenderIsAccountsModule returns true if the sender of the execution request is the accounts module.
func SenderIsAccountsModule(ctx context.Context) bool {
	return bytes.Equal(Sender(ctx), accountsModuleAddress)
}

// ExecModule can be used to execute a message towards a module.
func ExecModule[Resp any, RespProto implementation.ProtoMsgG[Resp], Req any, ReqProto implementation.ProtoMsgG[Req]](ctx context.Context, msg ReqProto) (RespProto, error) {
	return implementation.ExecModule[Resp, RespProto, Req, ReqProto](ctx, msg)
}

func ExecModuleUntyped(ctx context.Context, msg implementation.ProtoMsg) (implementation.ProtoMsg, error) {
	return implementation.ExecModuleUntyped(ctx, msg)
}

// QueryModule can be used by an account to execute a module query.
func QueryModule[Resp any, RespProto implementation.ProtoMsgG[Resp], Req any, ReqProto implementation.ProtoMsgG[Req]](ctx context.Context, req ReqProto) (RespProto, error) {
	return implementation.QueryModule[Resp, RespProto, Req, ReqProto](ctx, req)
}

// UnpackAny unpacks a protobuf Any message generically.
func UnpackAny[Msg any, ProtoMsg implementation.ProtoMsgG[Msg]](any *implementation.Any) (*Msg, error) {
	return implementation.UnpackAny[Msg, ProtoMsg](any)
}

// PackAny packs a protobuf Any message generically.
func PackAny(msg implementation.ProtoMsg) (*implementation.Any, error) {
	return implementation.PackAny(msg)
}

// ExecModuleAnys can be used to execute a list of messages towards a module
// when those messages are packed in Any messages. The function returns a list
// of responses packed in Any messages.
func ExecModuleAnys(ctx context.Context, msgs []*implementation.Any) ([]*implementation.Any, error) {
	responses := make([]*implementation.Any, len(msgs))
	for i, msg := range msgs {
		concreteMessage, err := implementation.UnpackAnyRaw(msg)
		if err != nil {
			return nil, fmt.Errorf("error unpacking message %d: %w", i, err)
		}
		resp, err := ExecModuleUntyped(ctx, concreteMessage)
		if err != nil {
			return nil, fmt.Errorf("error executing message %d: %w", i, err)
		}
		// pack again
		respAnyPB, err := implementation.PackAny(resp)
		if err != nil {
			return nil, fmt.Errorf("error packing response %d: %w", i, err)
		}
		responses[i] = respAnyPB
	}
	return responses, nil
}
