package runtime

import (
	"context"
	"fmt"

	"cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	"google.golang.org/protobuf/proto"
)

var _ appmodule.MsgRouter = (*handlerRouter)(nil)

func newRouters() (*handlerRouter, *preHandlerRouter, *postHandlerRouter, *handlerRouter) {
	return &handlerRouter{
			handlers: map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error){},
		},
		&preHandlerRouter{
			specificPreHandler: map[string][]func(ctx context.Context, msg proto.Message) error{},
		},
		&postHandlerRouter{
			specificPostHandler: map[string][]func(ctx context.Context, msg proto.Message, resp proto.Message) error{},
		},
		&handlerRouter{
			handlers: map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error){},
		}
}

type handlerRouter struct {
	err      error
	handlers map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error)
}

func (h *handlerRouter) Register(msgName string, handler appmodule.Handler) {
	if h.handlers != nil {
		return
	}
	if _, exist := h.handlers[msgName]; exist {
		h.err = fmt.Errorf("conflicting msg handlers: %s", msgName)
	}
	h.handlers[msgName] = handler
}

func (h *handlerRouter) build(pre *preHandlerRouter, post *postHandlerRouter) (
	func(ctx context.Context, msg proto.Message) (proto.Message, error),
	error,
) {
	handlers := make(map[string]func(ctx context.Context, msg proto.Message) (proto.Message, error))

	globalPreHandler := func(ctx context.Context, msg transaction.Type) error {
		for _, h := range pre.globalPreHandler {
			err := h(ctx, msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	globalPostHandler := func(ctx context.Context, msg, msgResp transaction.Type) error {
		for _, h := range post.globalPostHandler {
			err := h(ctx, msg, msgResp)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for msgType, handler := range h.handlers {
		// find pre handler
		specificPreHandlers := pre.specificPreHandler[msgType]
		// find post handler
		specificPostHandlers := post.specificPostHandler[msgType]
		// build the handler
		handlers[msgType] = buildHandler(handler, specificPreHandlers, globalPreHandler, specificPostHandlers, globalPostHandler)
	}
	// TODO: add checks for when a pre handler/post handler is registered but there is no matching handler.

	// return handler as function
	return func(ctx context.Context, msg transaction.Type) (transaction.Type, error) {
		typeName := proto.MessageName(msg)
		handler, exists := handlers[string(typeName)]
		if !exists {
			return nil, fmt.Errorf("%w: %s", appmodule.ErrNoHandler, typeName)
		}
		return handler(ctx, msg)
	}, nil
}

func buildHandler(
	handler func(ctx context.Context, msg proto.Message) (proto.Message, error),
	preHandlers []func(ctx context.Context, msg proto.Message) error, globalPreHandler func(ctx context.Context, msg proto.Message) error,
	postHandlers []func(ctx context.Context, msg proto.Message, resp proto.Message) error, globalPostHandler func(ctx context.Context, msg proto.Message, resp proto.Message) error) func(ctx context.Context, msg proto.Message) (proto.Message, error) {
	return func(ctx context.Context, msg proto.Message) (msgResp proto.Message, err error) {
		for _, preHandler := range preHandlers {
			if err := preHandler(ctx, msg); err != nil {
				return nil, err
			}
		}

		err = globalPreHandler(ctx, msg)
		if err != nil {
			return nil, err
		}

		msgResp, err = handler(ctx, msg)
		if err != nil {
			return nil, err
		}

		for _, postHandler := range postHandlers {
			if err := postHandler(ctx, msg, msgResp); err != nil {
				return nil, err
			}
		}

		err = globalPostHandler(ctx, msg, msgResp)
		return msgResp, err
	}
}

var _ appmodule.PreMsgRouter = (*preHandlerRouter)(nil)

type preHandlerRouter struct {
	err error

	specificPreHandler map[string][]func(ctx context.Context, msg proto.Message) error
	globalPreHandler   []func(ctx context.Context, msg proto.Message) error
}

func (p *preHandlerRouter) Register(msgName string, handler appmodule.PreMsgHandler) {
	p.specificPreHandler[msgName] = append(p.specificPreHandler[msgName], handler)
}

func (p *preHandlerRouter) RegisterGlobal(handler appmodule.PreMsgHandler) {
	p.globalPreHandler = append(p.globalPreHandler, handler)
}

var _ appmodule.PostMsgRouter = (*postHandlerRouter)(nil)

type postHandlerRouter struct {
	err error

	specificPostHandler map[string][]func(ctx context.Context, msg proto.Message, resp proto.Message) error
	globalPostHandler   []func(ctx context.Context, msg proto.Message, resp proto.Message) error
}

func (p postHandlerRouter) Register(msgName string, handler appmodule.PostMsgHandler) {
	p.specificPostHandler[msgName] = append(p.specificPostHandler[msgName], handler)
}

func (p postHandlerRouter) RegisterGlobal(handler appmodule.PostMsgHandler) {
	p.globalPostHandler = append(p.globalPostHandler, handler)
}
