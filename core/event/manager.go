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
	// EmitProto emits events represented as a protobuf message (as described in ADR 032).
	EmitProto(event protoiface.MessageV1) error

	// EmitKV emits an event based on an event and kv-pair attributes.
	EmitKV(eventType string, attrs ...KVEventAttribute) error
}

// KVEventAttribute is a kv-pair event attribute.
type KVEventAttribute struct {
	Key, Value string
}
