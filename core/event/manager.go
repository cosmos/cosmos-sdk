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
	// EmitProtoEvent emits events represented as a protobuf message (as described in ADR 032).
	EmitProtoEvent(ctx context.Context, event protoiface.MessageV1) error

	// EmitKVEvent emits an event based on an event and kv-pair attributes.
	EmitKVEvent(ctx context.Context, eventType string, attrs ...KVEventAttribute) error
}

// KVEventAttribute is a kv-pair event attribute.
type KVEventAttribute struct {
	Key, Value string
}
