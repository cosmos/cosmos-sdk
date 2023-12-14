package appmanager

import (
	"context"

	"github.com/cosmos/cosmos-sdk/serverv2/appmanager"
	"google.golang.org/protobuf/proto"
)

type Type = proto.Message
type Identity = []byte

type Tx = appmanager.Tx

type MsgRouterBuilder interface {
	RegisterHandler(msg Type, handlerFunc func(ctx context.Context, msg Type) (resp Type, err error))
}

type QueryRouterBuilder = MsgRouterBuilder

type PreMsgRouterBuilder interface {
	RegisterPreHandler(msg Type, preHandler func(ctx context.Context, msg Type) error)
}

type PostMsgRouterBuilder interface {
	RegisterPostHandler(msg Type, postHandler func(ctx context.Context, msg Type, msgResp Type) error)
}

type STFModule interface {
	Name() string
	RegisterMsgHandlers(router MsgRouterBuilder)
	RegisterQueryHandler(router QueryRouterBuilder)
	BeginBlocker() func(ctx context.Context) error
	EndBlocker() func(ctx context.Context) error
	TxValidator() func(ctx context.Context, tx Tx) error
	RegisterPreMsgHandler(router PreMsgRouterBuilder)
	RegisterPostMsgHandler(router PostMsgRouterBuilder)
}

type Module interface {
	STFModule
}
