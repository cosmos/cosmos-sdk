package stf

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/core/transaction"
)

var ErrNoHandler = errors.New("no handler")

// MsgHandler is a function that handles the message execution.
// TODO: move to appmanager.Module.go (marko)
type MsgHandler = func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error)

// PreMsgHandler is a function that executes before the message execution.
// TODO: move to appmanager.Module.go (marko)
type PreMsgHandler = func(ctx context.Context, msg transaction.Type) (err error)

// PostMsgHandler is a function that executes after the message execution.
// TODO: move to appmanager.Module.go (marko)
type PostMsgHandler = func(ctx context.Context, msg, msgResp transaction.Type) (err error)

type QueryHandler = MsgHandler

// TODO: make a case for *, listen to all messages

// NewMsgRouterBuilder is a router that routes messages to their respective handlers.
func NewMsgRouterBuilder() *MsgRouterBuilder {
	return &MsgRouterBuilder{
		handlers:     make(map[string]MsgHandler),
		preHandlers:  make(map[string][]PreMsgHandler),
		postHandlers: make(map[string][]PostMsgHandler),
	}
}

type MsgRouterBuilder struct {
	handlers           map[string]MsgHandler
	globalPreHandlers  []PreMsgHandler
	preHandlers        map[string][]PreMsgHandler // TODO document how to do ordering, if needed
	postHandlers       map[string][]PostMsgHandler
	globalPostHandlers []PostMsgHandler
}

func (b *MsgRouterBuilder) RegisterHandler(msgType string, handler MsgHandler) error {
	// panic on override
	if _, ok := b.handlers[msgType]; ok {
		return fmt.Errorf("handler already registered: %s", msgType)
	}
	b.handlers[msgType] = handler
	return nil
}

func (b *MsgRouterBuilder) RegisterGlobalPreHandler(handler PreMsgHandler) {
	b.globalPreHandlers = append(b.globalPreHandlers, handler)
}

func (b *MsgRouterBuilder) RegisterPreHandler(msgType string, handler PreMsgHandler) {
	b.preHandlers[msgType] = append(b.preHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterPostHandler(msgType string, handler PostMsgHandler) {
	b.postHandlers[msgType] = append(b.postHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterGlobalPostHandler(handler PostMsgHandler) {
	b.globalPostHandlers = append(b.globalPostHandlers, handler)
}

func (b *MsgRouterBuilder) Build() (MsgHandler, error) {
	handlers := make(map[string]MsgHandler)

	globalPreHandler := func(ctx context.Context, msg transaction.Type) error {
		for _, h := range b.globalPreHandlers {
			err := h(ctx, msg)
			if err != nil {
				return err
			}
		}
		return nil
	}

	globalPostHandler := func(ctx context.Context, msg, msgResp transaction.Type) error {
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
	// TODO: add checks for when a pre handler/post handler is registered but there is no matching handler.

	// return handler as function
	return func(ctx context.Context, msg transaction.Type) (transaction.Type, error) {
		typeName := typeName(msg)
		handler, exists := handlers[typeName]
		if !exists {
			return nil, fmt.Errorf("%w: %s", ErrNoHandler, typeName)
		}
		return handler(ctx, msg)
	}, nil
}

func buildHandler(handler MsgHandler, preHandlers []PreMsgHandler, globalPreHandler PreMsgHandler, postHandlers []PostMsgHandler, globalPostHandler PostMsgHandler) MsgHandler {
	return func(ctx context.Context, msg transaction.Type) (msgResp transaction.Type, err error) {
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
