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
	//
	// Callers SHOULD assume that these events may be included in consensus. These events
	// MUST be emitted deterministically and adding, removing or changing these events SHOULD
	// be considered state-machine breaking.
	EmitProtoEvent(ctx context.Context, event protoiface.MessageV1) error

	// EmitKVEvent emits an event based on an event and kv-pair attributes.
	//
	// These events will not be part of consensus and adding, removing or changing these events is
	// not a state-machine breaking change.
	EmitKVEvent(ctx context.Context, eventType string, attrs ...KVEventAttribute) error

	// EmitProtoEventNonConsensus emits events represented as a protobuf message (as described in ADR 032), without
	// including it in blockchain consensus.
	//
	// These events will not be part of consensus and adding, removing or changing events is
	// not a state-machine breaking change.
	EmitProtoEventNonConsensus(ctx context.Context, event protoiface.MessageV1) error
}

// KVEventAttribute is a kv-pair event attribute.
type KVEventAttribute struct {
	Key, Value string
}
