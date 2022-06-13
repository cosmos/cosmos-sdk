package event

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Service represents an event service which can retrieve and set an event manager in a context.
// event.Service is a core API type that should be provided by the runtime module being used to
// build an app via depinject.
type Service interface {
	// GetManager returns the event manager for the context or a no-op event manager if one isn't attached.
	GetManager(context.Context) Manager

	// WithManager returns a new context with the provided event manager attached.
	WithManager(context.Context, Manager) context.Context
}

// Manager represents an event manager.
type Manager interface {

	// Emit emits a typed protobuf event.
	Emit(proto.Message) error

	// EmitLegacy emits a legacy (untyped) tendermint event.
	EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error
}

// LegacyEventAttribute is a legacy (untyped) event attribute.
type LegacyEventAttribute struct {
	Key, Value string
}
