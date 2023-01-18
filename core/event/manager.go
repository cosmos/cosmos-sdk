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
	// GetEventManager returns the event manager for the context or a no-op event manager if one isn't attached.
	GetEventManager(context.Context) Manager
}

// Manager represents an event manager.
type Manager interface {
	// Emit emits typed events to both clients and state machine listeners
	// (if supported). These events MUST be emitted deterministically
	// and should be assumed to be part of blockchain consensus.
	Emit(protoiface.MessageV1) error

	// EmitLegacy emits legacy untyped events to clients only. These events do not need to be emitted deterministically
	// and are not part of blockchain consensus.
	EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error

	// EmitClientOnly emits events only to clients. These events do not need to be emitted deterministically
	// and are not part of blockchain consensus.
	EmitClientOnly(protoiface.MessageV1) error
}

// LegacyEventAttribute is a legacy (untyped) event attribute.
type LegacyEventAttribute struct {
	Key, Value string
}
