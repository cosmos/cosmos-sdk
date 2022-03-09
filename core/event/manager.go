package event

import (
	"context"

	"google.golang.org/protobuf/proto"
)

// Manager represents an event manager.
type Manager interface {
	Emit(proto.Message) error
	EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error
}

// LegacyEventAttribute is a legacy (untyped) event attribute.
type LegacyEventAttribute struct {
	Key, Value string
}

// GetManager will always return a non-nil event manager with a no-op event
// manager being returned if there is no manager in the context.
func GetManager(ctx context.Context) Manager {
	panic("TODO")
}

// WithManager creates a new context with the provided event manager.
func WithManager(ctx context.Context, manager Manager) context.Context {
	panic("TODO")
}
