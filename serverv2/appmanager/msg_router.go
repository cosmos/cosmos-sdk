package appmanager

import (
	"context"
	"errors"
)

var (
	ErrNoHandler = errors.New("no handler")
)

// MsgHandler is a function that handles the message execution.
type MsgHandler = func(ctx context.Context, msg Type) (msgResp Type, err error)

// PreMsgHandler is a function that executes before the message execution.
type PreMsgHandler = func(ctx context.Context, msg Type) (err error)

// PostMsgHandler is a function that executes after the message execution.
type PostMsgHandler = func(ctx context.Context, msg Type, msgResp Type) (err error)

type MsgRouter struct {
	handlers map[string]MsgHandler
}

func (r MsgRouter) Handle(ctx context.Context, msg Type) (msgResp Type, err error) {
	msgType := TypeName(msg)
	handler, ok := r.handlers[msgType]
	if !ok {
		return nil, ErrNoHandler
	}
	return handler(ctx, msg)
}

type MsgRouterBuilder struct {
	handlers     map[string]MsgHandler
	preHandlers  map[string][]PreMsgHandler
	postHandlers map[string][]PostMsgHandler
}

func (b *MsgRouterBuilder) RegisterHandler(msgType string, handler MsgHandler) {
	// panic on override
	if _, ok := b.handlers[msgType]; ok {
		panic("handler already registered: " + msgType)
	}
	b.handlers[msgType] = handler
}

func (b *MsgRouterBuilder) RegisterPreHandler(msgType string, handler PreMsgHandler) {
	b.preHandlers[msgType] = append(b.preHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) RegisterPostHandler(msgType string, handler PostMsgHandler) {
	b.postHandlers[msgType] = append(b.postHandlers[msgType], handler)
}

func (b *MsgRouterBuilder) Build() (*MsgRouter, error) {
	msgRouter := &MsgRouter{
		handlers: make(map[string]MsgHandler),
	}
	for msgType, handler := range b.handlers {
		// find pre handler
		preHandlers := b.preHandlers[msgType]
		// find post handler
		postHandlers := b.postHandlers[msgType]
		// build the handler
		msgRouter.handlers[msgType] = buildHandler(handler, preHandlers, postHandlers)
	}
	// TODO: add checks for when a pre handler/post handler is registered but there is no matching handler.
	return msgRouter, nil
}

func buildHandler(handler MsgHandler, preHandlers []PreMsgHandler, postHandlers []PostMsgHandler) MsgHandler {
	// TODO: maybe we can optimize this by doing a switch case and checking if the pre/post handlers are empty
	// in order to avoid pointless iterations when there are no pre/post handlers
	return func(ctx context.Context, msg Type) (msgResp Type, err error) {
		for _, preHandler := range preHandlers {
			if err := preHandler(ctx, msg); err != nil {
				return nil, err
			}
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
		return msgResp, nil
	}
}
