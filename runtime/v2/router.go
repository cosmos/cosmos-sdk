package runtime

import (
	"fmt"

	"cosmossdk.io/core/appmodule/v2"
)

var _ appmodule.MsgRouter = (*handlersBuilder)(nil)

type handlersBuilder struct {
	err        error
	v1Handlers map[string]v1Handler
	v2Handlers map[string]v2Handler
}

func (h *handlersBuilder) Register(msgName string, handler any) {
	switch handler := handler.(type) {
	case v1Handler:
		insertIfNotExistOr(h.v1Handlers, msgName, handler, func() {
			h.err = fmt.Errorf("v1 handler for msg '%s' already registered", msgName)
		})
	case v2Handler:
		insertIfNotExistOr(h.v2Handlers, msgName, handler, func() {
			h.err = fmt.Errorf("v2 handler for msg '%s' already registered", msgName)
		})
	default:
		h.err = fmt.Errorf("invalid handler type: %T", handler)
	}
}

func insertIfNotExistOr[K comparable, V any](m map[K]V, key K, value V, f func()) {
	_, exists := m[key]
	if exists {
		f()
		return
	}
	m[key] = value
}

var _ appmodule.PreMsgRouter = (*preHandlerBuilder)(nil)

type preHandlerBuilder struct {
	err error

	v1PreHandlers      map[string]v1PreHandler
	v1PreHandlerGlobal []v1PreHandler
	v2PreHandlers      map[string]v2PreHandler
	v2PreHandlerGlobal []v2PreHandler
}

func (p *preHandlerBuilder) Register(msgName string, handler any) {
	switch handler := handler.(type) {
	case v1PreHandler:
		p.v1PreHandlers[msgName] = handler
	case v2PreHandler:
		p.v2PreHandlers[msgName] = handler
	default:
		p.err = fmt.Errorf("invalid pre handler type: %T", handler)
	}
}

func (p *preHandlerBuilder) RegisterGlobal(handler any) {
	switch handler := handler.(type) {
	case v1PreHandler:
		p.v1PreHandlerGlobal = append(p.v1PreHandlerGlobal, handler)
	case v2PreHandler:
		p.v2PreHandlerGlobal = append(p.v2PreHandlerGlobal, handler)
	default:
		p.err = fmt.Errorf("invalid global pre handler type: %T", handler)
	}
}

type postHandlerBuilder struct {
	err error

	v1PostHandlers      map[string]v1PostHandler
	v1PostHandlerGlobal []v1PostHandler
	v2PostHandlers      map[string]v2PostHandler
	v2PostHandlerGlobal []v2PostHandler
}

func (p *postHandlerBuilder) Register(msgName string, handler any) {
	switch handler := handler.(type) {
	case v1PostHandler:
		p.v1PostHandlers[msgName] = handler
	case v2PostHandler:
		p.v2PostHandlers[msgName] = handler
	default:
		p.err = fmt.Errorf("invalid post handler type: %T", handler)
	}
}

func (p *postHandlerBuilder) RegisterGlobal(handler any) {
	switch handler := handler.(type) {
	case v1PostHandler:
		p.v1PostHandlerGlobal = append(p.v1PostHandlerGlobal, handler)
	case v2PostHandler:
		p.v2PostHandlerGlobal = append(p.v2PostHandlerGlobal, handler)
	default:
		p.err = fmt.Errorf("invalid global post handler type: %T", handler)
	}
}
