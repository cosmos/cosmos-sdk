package appmanager

import (
	"context"

	"cosmossdk.io/core/transaction"
)

type Tx = transaction.Tx

type MsgRouterBuilder interface {
	RegisterHandler(msg Type, handlerFunc func(ctx context.Context, msg Type) (resp Type, err error))
}

type QueryRouterBuilder = MsgRouterBuilder

type PreMsgRouterBuilder interface {
	RegisterPreHandler(msg Type, preHandler func(ctx context.Context, msg Type) error)
}

type PostMsgRouterBuilder interface {
	RegisterPostHandler(msg Type, postHandler func(ctx context.Context, msg, msgResp Type) error)
}
