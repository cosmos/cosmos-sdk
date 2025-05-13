package runtime

import (
	"context"

	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/event"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ event.Service = (*EventService)(nil)

type EventService struct {
	Events
}

func (es EventService) EventManager(ctx context.Context) event.Manager {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &Events{sdkCtx.EventManager()}
}

var _ event.Manager = (*Events)(nil)

type Events struct {
	sdk.EventManagerI
}

func NewEventManager(ctx context.Context) event.Manager {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &Events{sdkCtx.EventManager()}
}

// Emit emits an typed event that is defined in the protobuf file.
// In the future these events will be added to consensus.
func (e Events) Emit(_ context.Context, event protoiface.MessageV1) error {
	return e.EmitTypedEvent(event)
}

// EmitKV emits a key value pair event.
func (e Events) EmitKV(_ context.Context, eventType string, attrs ...event.Attribute) error {
	attributes := make([]sdk.Attribute, 0, len(attrs))

	for _, attr := range attrs {
		attributes = append(attributes, sdk.NewAttribute(attr.Key, attr.Value))
	}

	e.EmitEvents(sdk.Events{sdk.NewEvent(eventType, attributes...)})
	return nil
}

// EmitNonConsensus emits an typed event that is defined in the protobuf file.
// In the future these events will be added to consensus.
func (e Events) EmitNonConsensus(_ context.Context, event protoiface.MessageV1) error {
	return e.EmitTypedEvent(event)
}
