package stf

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	gogoproto "github.com/cosmos/gogoproto/proto"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/router"
)

var ErrNoHandler = errors.New("no handler")

// NewMsgRouterBuilder is a router that routes messages to their respective handlers.
func NewMsgRouterBuilder() *MsgRouterBuilder {
	return &MsgRouterBuilder{
		handlers:     make(map[string]appmodulev2.Handler),
		preHandlers:  make(map[string][]appmodulev2.PreMsgHandler),
		postHandlers: make(map[string][]appmodulev2.PostMsgHandler),
	}
}

type MsgRouterBuilder struct {
	handlers           map[string]appmodulev2.Handler
	globalPreHandlers  []appmodulev2.PreMsgHandler
	preHandlers        map[string][]appmodulev2.PreMsgHandler
	postHandlers       map[string][]appmodulev2.PostMsgHandler
	globalPostHandlers []appmodulev2.PostMsgHandler
}

func (b *MsgRouterBuilder) RegisterHandler(msgType string, handler appmodulev2.Handler) error {
	// panic on override
	if _, ok := b.handlers[msgType]; ok {
		return fmt.Errorf("handler already registered: %s", msgType)
	}
	b.handlers[msgType] = handler
	return nil
}

func (b *MsgRouterBuilder) RegisterGlobalPreHandler(handler appmodulev2.PreMsgHandler) {
	b.globalPreHandlers = append(b.globalPreHandlers, handler)
}

func (b *MsgRouterBuilder) RegisterPreHandler(msgType string, handler appmodulev2.PreMsgHandler) {
	b.preHandlers[msgType] = append(b.preHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterPostHandler(msgType string, handler appmodulev2.PostMsgHandler) {
	b.postHandlers[msgType] = append(b.postHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterGlobalPostHandler(handler appmodulev2.PostMsgHandler) {
	b.globalPostHandlers = append(b.globalPostHandlers, handler)
}

func (b *MsgRouterBuilder) HandlerExists(msgType string) bool {
	_, ok := b.handlers[msgType]
	return ok
}

func (b *MsgRouterBuilder) Build() (Router, error) {
	handlers := make(map[string]appmodulev2.Handler)

	globalPreHandler := func(ctx context.Context, msg appmodulev2.Message) error {
		for _, h := range b.globalPreHandlers {
			err := h(ctx, msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	globalPostHandler := func(ctx context.Context, msg, msgResp appmodulev2.Message) error {
		for _, h := range b.globalPostHandlers {
			err := h(ctx, msg, msgResp)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for msgType, handler := range b.handlers {
		// find pre handler
		preHandlers := b.preHandlers[msgType]
		// find post handler
		postHandlers := b.postHandlers[msgType]
		// build the handler
		handlers[msgType] = buildHandler(handler, preHandlers, globalPreHandler, postHandlers, globalPostHandler)
	}

	return Router{
		handlers: handlers,
	}, nil
}

func buildHandler(
	handler appmodulev2.Handler,
	preHandlers []appmodulev2.PreMsgHandler,
	globalPreHandler appmodulev2.PreMsgHandler,
	postHandlers []appmodulev2.PostMsgHandler,
	globalPostHandler appmodulev2.PostMsgHandler,
) appmodulev2.Handler {
	return func(ctx context.Context, msg appmodulev2.Message) (msgResp appmodulev2.Message, err error) {
		if len(preHandlers) != 0 {
			for _, preHandler := range preHandlers {
				if err := preHandler(ctx, msg); err != nil {
					return nil, err
				}
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

		if len(postHandlers) != 0 {
			for _, postHandler := range postHandlers {
				if err := postHandler(ctx, msg, msgResp); err != nil {
					return nil, err
				}
			}
		}
		err = globalPostHandler(ctx, msg, msgResp)
		return msgResp, err
	}
}

// msgTypeURL returns the TypeURL of a proto message.
func msgTypeURL(msg gogoproto.Message) string {
	return gogoproto.MessageName(msg)
}

var _ router.Service = (*Router)(nil)

// Router implements the STF router for msg and query handlers.
type Router struct {
	handlers map[string]appmodulev2.Handler
}

func (r Router) CanInvoke(_ context.Context, typeURL string) error {
	_, exists := r.handlers[typeURL]
	if !exists {
		return fmt.Errorf("%w: %s", ErrNoHandler, typeURL)
	}
	return nil
}

func (r Router) InvokeTyped(ctx context.Context, req, resp gogoproto.Message) error {
	handlerResp, err := r.InvokeUntyped(ctx, req)
	if err != nil {
		return err
	}
	merge(handlerResp, resp)
	return nil
}

func merge(src, dst gogoproto.Message) {
	reflect.Indirect(reflect.ValueOf(dst)).Set(reflect.Indirect(reflect.ValueOf(src)))
}

func (r Router) InvokeUntyped(ctx context.Context, req gogoproto.Message) (res gogoproto.Message, err error) {
	typeName := msgTypeURL(req)
	fmt.Println("typeName", typeName)
	handler, exists := r.handlers[typeName]
	fmt.Println("handler", exists)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNoHandler, typeName)
	}
	return handler(ctx, req)
}
