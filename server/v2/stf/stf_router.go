package stf

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
)

var ErrNoHandler = errors.New("no handler")

// NewMsgRouterBuilder is a router that routes messages to their respective handlers.
func NewMsgRouterBuilder() *MsgRouterBuilder {
	return &MsgRouterBuilder{
		handlers:     make(map[string]appmodulev2.HandlerFunc),
		preHandlers:  make(map[string][]appmodulev2.PreMsgHandler),
		postHandlers: make(map[string][]appmodulev2.PostMsgHandler),
	}
}

type MsgRouterBuilder struct {
	handlers           map[string]appmodulev2.HandlerFunc
	globalPreHandlers  []appmodulev2.PreMsgHandler
	preHandlers        map[string][]appmodulev2.PreMsgHandler
	postHandlers       map[string][]appmodulev2.PostMsgHandler
	globalPostHandlers []appmodulev2.PostMsgHandler
}

func (b *MsgRouterBuilder) RegisterHandler(msgType string, handler appmodulev2.HandlerFunc) error {
	// panic on override
	if _, ok := b.handlers[msgType]; ok {
		return fmt.Errorf("handler already registered: %s", msgType)
	}
	b.handlers[msgType] = handler
	return nil
}

func (b *MsgRouterBuilder) RegisterGlobalPreMsgHandler(handler appmodulev2.PreMsgHandler) {
	b.globalPreHandlers = append(b.globalPreHandlers, handler)
}

func (b *MsgRouterBuilder) RegisterPreMsgHandler(msgType string, handler appmodulev2.PreMsgHandler) {
	b.preHandlers[msgType] = append(b.preHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterPostMsgHandler(msgType string, handler appmodulev2.PostMsgHandler) {
	b.postHandlers[msgType] = append(b.postHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterGlobalPostMsgHandler(handler appmodulev2.PostMsgHandler) {
	b.globalPostHandlers = append(b.globalPostHandlers, handler)
}

func (b *MsgRouterBuilder) HandlerExists(msgType string) bool {
	_, ok := b.handlers[msgType]
	return ok
}

func (b *MsgRouterBuilder) build() (coreRouterImpl, error) {
	handlers := make(map[string]appmodulev2.HandlerFunc)

	globalPreHandler := func(ctx context.Context, msg transaction.Msg) error {
		for _, h := range b.globalPreHandlers {
			err := h(ctx, msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	globalPostHandler := func(ctx context.Context, msg, msgResp transaction.Msg) error {
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

	return coreRouterImpl{
		handlers: handlers,
	}, nil
}

func buildHandler(
	handler appmodulev2.HandlerFunc,
	preHandlers []appmodulev2.PreMsgHandler,
	globalPreHandler appmodulev2.PreMsgHandler,
	postHandlers []appmodulev2.PostMsgHandler,
	globalPostHandler appmodulev2.PostMsgHandler,
) appmodulev2.HandlerFunc {
	return func(ctx context.Context, msg transaction.Msg) (msgResp transaction.Msg, err error) {
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

var _ router.Service = (*coreRouterImpl)(nil)

// coreRouterImpl implements the STF router for msg and query handlers.
type coreRouterImpl struct {
	handlers map[string]appmodulev2.HandlerFunc
}

func (r coreRouterImpl) CanInvoke(_ context.Context, typeURL string) error {
	// trimming prefixes is a backwards compatibility strategy that we use
	// for baseapp components that did routing through type URL rather
	// than protobuf message names.
	typeURL = strings.TrimPrefix(typeURL, "/")
	_, exists := r.handlers[typeURL]
	if !exists {
		return fmt.Errorf("%w: %s", ErrNoHandler, typeURL)
	}
	return nil
}

func (r coreRouterImpl) Invoke(ctx context.Context, req transaction.Msg) (res transaction.Msg, err error) {
	typeName := msgTypeURL(req)
	handler, exists := r.handlers[typeName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNoHandler, typeName)
	}

	return handler(ctx, req)
}
