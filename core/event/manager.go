package event

import (
	"context"
	"google.golang.org/protobuf/proto"
)

type Manager interface {
	Emit(proto.Message) error
	EmitLegacy(eventType string, attrs ...LegacyEventAttribute) error
}

type LegacyEventAttribute struct {
	Key, Value string
}

func GetManager(ctx context.Context) Manager {
	panic("TODO")
}

func WithManager(ctx context.Context, manager Manager) context.Context {
	panic("TODO")
}
