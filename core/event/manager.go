// Package event provides a basic API for app modules to emit events.
package event

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"
)

// Service represents an event service which can retrieve and set an event manager in a context.
// event.Service is a core API type that should be provided by the runtime module being used to
// build an app via depinject.
type Service interface {
	// GetManager returns the event manager for the context or a no-op event manager if one isn't attached.
	GetManager(context.Context) Manager
}

// Manager represents an event manager.
type Manager interface {

	// Emit emits a typed protobuf event.
	Emit(protoiface.MessageV1) error

	// EmitLegacy emits a legacy (untyped) tendermint event.
	EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error
}

// LegacyEventAttribute is a legacy (untyped) event attribute.
type LegacyEventAttribute struct {
	Key, Value string
}
