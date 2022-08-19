package appmodule

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// EventListener is an empty interface to indicate an event listener type.
type EventListener interface{}

// AddEventListener adds an event listener for the provided type to the handler.
func AddEventListener[E protoiface.MessageV1](h *Handler, listener func(ctx context.Context, e E)) {
	h.EventListeners = append(h.EventListeners, listener)
}

// AddEventInterceptor adds an event interceptor for the provided type to the handler which can
// return an error to interrupt state machine processing and revert uncommitted state changes.
func AddEventInterceptor[E protoiface.MessageV1](h *Handler, interceptor func(ctx context.Context, e E) error) {
	h.EventListeners = append(h.EventListeners, interceptor)
}
