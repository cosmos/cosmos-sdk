package sdk

import (
	"context"

	internalaccounts "cosmossdk.io/x/accounts/internal/accounts"
	"github.com/cosmos/gogoproto/proto"
)

type Msg[T any] interface {
	*T
	proto.Message
}

type (
	QueryRouter       = internalaccounts.QueryRouter
	ExecuteRouter     = internalaccounts.ExecuteRouter
	InitRouter        = internalaccounts.InitRouter
	Account           = internalaccounts.Account
	BuildDependencies = internalaccounts.BuildDependencies
)

func RegisterQueryHandler[
	Req any, Resp any, ReqP Msg[Req], RespP Msg[Resp],
](router *QueryRouter, handler func(ctx context.Context, msg Req) (Resp, error)) {
	internalaccounts.RegisterQueryHandler[Req, Resp, ReqP, RespP](router, handler)
}

func RegisterExecuteHandler[
	Req any, Resp any, ReqP Msg[Req], RespP Msg[Resp],
](router *ExecuteRouter, handler func(ctx context.Context, msg Req) (Resp, error)) {
	internalaccounts.RegisterExecuteHandler[Req, Resp, ReqP, RespP](router, handler)
}

func RegisterInitHandler[Req, Resp any, ReqP Msg[Req], RespP Msg[Resp]](router *InitRouter, handler func(ctx context.Context, msg Req) (Resp, error)) {
	internalaccounts.RegisterInitHandler[Req, Resp, ReqP, RespP](router, handler)
}

func Sender(ctx context.Context) []byte { return internalaccounts.Sender(ctx) }

func Whoami(ctx context.Context) []byte { return internalaccounts.Whoami(ctx) }
