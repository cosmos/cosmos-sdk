package appmodule

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Handler is the interface that wraps a handler for modules to register state transitions
type MsgRouterBuilder interface {
	RegisterHandler(msg proto.Message, handlerFunc func(ctx context.Context, msg proto.Message) (resp proto.Message, err error))
}

// QueryRouterBuilder is the interface that wraps a handler for modules to register query logc
type QueryRouterBuilder = MsgRouterBuilder

// PreMsgRouterBuilder is the interface that wraps a handler for modules to register logic to be run post message execution
type PreMsgRouterBuilder interface {
	RegisterPreHandler(msg proto.Message, preHandler func(ctx context.Context, msg proto.Message) error)
}

// PostMsgRouterBuilder is the interface that wraps a handler for modules to register logic to be run post message execution.
type PostMsgRouterBuilder interface {
	RegisterPostHandler(msg proto.Message, postHandler func(ctx context.Context, msg, msgResp proto.Message) error)
}
