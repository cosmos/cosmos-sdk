package coretesting

import (
	"reflect"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/core/event"
)

func TestEventsService(t *testing.T) {
	ctx := Context()
	es := EventsService(ctx, "auth")

	wantProtoEvent := &gogotypes.BoolValue{Value: true}
	err := es.EventManager(ctx).Emit(wantProtoEvent)
	if err != nil {
		t.Errorf("failed to emit proto event: %s", err)
	}

	wantEvent := event.NewEvent("new-account", event.Attribute{
		Key:   "number",
		Value: "1",
	})
	err = es.EventManager(ctx).EmitKV(wantEvent.Type, wantEvent.Attributes...)
	if err != nil {
		t.Errorf("failed to emit event: %s", err)
	}

	gotProtoEvents := es.GetProtoEvents(ctx)
	if len(gotProtoEvents) != 1 || gotProtoEvents[0] != wantProtoEvent {
		t.Errorf("unexpected proto events: got %v, want %v", gotProtoEvents, wantProtoEvent)
	}

	gotEvents := es.GetEvents(ctx)
	if len(gotEvents) != 1 || !reflect.DeepEqual(gotEvents[0], wantEvent) {
		t.Errorf("unexpected events: got %v, want %v", gotEvents, wantEvent)
	}

	// test reset
	es = EventsService(ctx, "auth")
	if es.GetEvents(ctx) != nil {
		t.Errorf("expected nil events after reset, got %v", es.GetEvents(ctx))
	}
	if es.GetProtoEvents(ctx) != nil {
		t.Errorf("expected nil proto events after reset, got %v", es.GetProtoEvents(ctx))
	}
}
