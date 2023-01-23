package appmodule

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// HasEventListeners is the extension interface that modules should implement to register
// event listeners.
type HasEventListeners interface {
	AppModule

	// RegisterEventListeners registers the module's events listeners.
	RegisterEventListeners(registrar *EventListenerRegistrar)
}

// EventListenerRegistrar allows registering event listeners.
type EventListenerRegistrar struct {
	listeners []any
}

// GetListeners gets the event listeners that have been registered
func (e *EventListenerRegistrar) GetListeners() []any {
	return e.listeners
}

// RegisterEventListener registers an event listener for event type E.
func RegisterEventListener[E protoiface.MessageV1](registrar *EventListenerRegistrar, listener func(context.Context, E)) {
	registrar.listeners = append(registrar.listeners, listener)
}

// RegisterEventInterceptor registers an event interceptor for event type E. Event interceptors can return errors
// to cause the process which emitted the event to fail.
func RegisterEventInterceptor[E protoiface.MessageV1](registrar *EventListenerRegistrar, interceptor func(context.Context, E) error) {
	registrar.listeners = append(registrar.listeners, interceptor)
}
