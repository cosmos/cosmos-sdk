package coretesting

import (
	"testing"

	"cosmossdk.io/core/event"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestEventsService(t *testing.T) {
	ctx := Context()
	es := EventsService(ctx, "auth")

	wantProtoEvent := &wrapperspb.BoolValue{Value: true}
	err := es.EventManager(ctx).Emit(wantProtoEvent)
	require.NoError(t, err)

	wantEvent := event.NewEvent("new-account", event.Attribute{
		Key:   "number",
		Value: "1",
	})
	err = es.EventManager(ctx).EmitKV(wantEvent.Type, wantEvent.Attributes...)
	require.NoError(t, err)

	gotProtoEvents := es.GetProtoEvents(ctx)
	require.Equal(t, []protoiface.MessageV1{wantProtoEvent}, gotProtoEvents)

	gotEvents := es.GetEvents(ctx)
	require.Equal(t, []event.Event{wantEvent}, gotEvents)

	// test reset
	es = EventsService(ctx, "auth")
	require.Nil(t, es.GetEvents(ctx))
	require.Nil(t, es.GetProtoEvents(ctx))
}
