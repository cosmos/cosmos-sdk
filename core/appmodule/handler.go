package appmodule

import (
	"context"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/core/transaction"
)

type Tx = transaction.Tx

type MsgRouterBuilder interface {
	RegisterHandler(msg proto.Message, handlerFunc func(ctx context.Context, msg proto.Message) (resp proto.Message, err error))
}

type QueryRouterBuilder = MsgRouterBuilder

type PreMsgRouterBuilder interface {
	RegisterPreHandler(msg proto.Message, preHandler func(ctx context.Context, msg proto.Message) error)
}

type PostMsgRouterBuilder interface {
	RegisterPostHandler(msg proto.Message, postHandler func(ctx context.Context, msg, msgResp proto.Message) error)
}
