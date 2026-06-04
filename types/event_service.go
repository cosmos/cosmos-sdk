package types

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/event"
)

var _ event.Service = (*EventService)(nil)

// EventService implements cosmossdk.io/core/event.Service via the SDK context.
type EventService struct {
	eventAdapter
}

func (es EventService) EventManager(ctx context.Context) event.Manager {
	return &eventAdapter{UnwrapSDKContext(ctx).EventManager()}
}

var _ event.Manager = (*eventAdapter)(nil)

type eventAdapter struct {
	*EventManager
}

func (e eventAdapter) Emit(_ context.Context, ev protoiface.MessageV1) error {
	return e.EmitTypedEvent(ev)
}

func (e eventAdapter) EmitKV(_ context.Context, eventType string, attrs ...event.Attribute) error {
	attributes := make([]Attribute, 0, len(attrs))
	for _, attr := range attrs {
		attributes = append(attributes, NewAttribute(attr.Key, attr.Value))
	}
	e.EmitEvents(Events{NewEvent(eventType, attributes...)})
	return nil
}

func (e eventAdapter) EmitNonConsensus(_ context.Context, ev protoiface.MessageV1) error {
	return e.EmitTypedEvent(ev)
}
