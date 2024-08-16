package stf

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	gogoproto "github.com/cosmos/gogoproto/proto"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/router"
	transaction "cosmossdk.io/core/transaction"
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

func (b *MsgRouterBuilder) Build() (coreRouterImpl, error) {
	handlers := make(map[string]appmodulev2.Handler)

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
	handler appmodulev2.Handler,
	preHandlers []appmodulev2.PreMsgHandler,
	globalPreHandler appmodulev2.PreMsgHandler,
	postHandlers []appmodulev2.PostMsgHandler,
	globalPostHandler appmodulev2.PostMsgHandler,
) appmodulev2.Handler {
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
	handlers map[string]appmodulev2.Handler
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

func (r coreRouterImpl) InvokeTyped(ctx context.Context, req, resp transaction.Msg) error {
	handlerResp, err := r.InvokeUntyped(ctx, req)
	if err != nil {
		return err
	}
	return merge(handlerResp, resp)
}

func (r coreRouterImpl) InvokeUntyped(ctx context.Context, req transaction.Msg) (res transaction.Msg, err error) {
	typeName := msgTypeURL(req)
	handler, exists := r.handlers[typeName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrNoHandler, typeName)
	}
	return handler(ctx, req)
}

// merge merges together two protobuf messages by setting the pointer
// to src in dst. Used internally.
func merge(src, dst gogoproto.Message) error {
	if src == nil {
		return fmt.Errorf("source message is nil")
	}
	if dst == nil {
		return fmt.Errorf("destination message is nil")
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() == reflect.Interface {
		srcVal = srcVal.Elem()
	}
	if dstVal.Kind() == reflect.Interface {
		dstVal = dstVal.Elem()
	}

	if srcVal.Kind() != reflect.Ptr || dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("both source and destination must be pointers")
	}

	srcElem := srcVal.Elem()
	dstElem := dstVal.Elem()

	if !srcElem.Type().AssignableTo(dstElem.Type()) {
		return fmt.Errorf("incompatible types: cannot merge %v into %v", srcElem.Type(), dstElem.Type())
	}

	dstElem.Set(srcElem)
	return nil
}
