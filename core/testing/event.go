package coretesting

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/event"
)

var _ event.Service = (*MemEventsService)(nil)

// EventsService attaches an event service to the context.
// Adding an existing module will reset the events.
func EventsService(ctx context.Context, moduleName string) MemEventsService {
	unwrap(ctx).events[moduleName] = nil
	unwrap(ctx).protoEvents[moduleName] = nil
	return MemEventsService{moduleName: moduleName}
}

type MemEventsService struct {
	moduleName string
}

func (e MemEventsService) EventManager(ctx context.Context) event.Manager {
	return eventManager{moduleName: e.moduleName, ctx: unwrap(ctx)}
}

func (e MemEventsService) GetEvents(ctx context.Context) []event.Event {
	return unwrap(ctx).events[e.moduleName]
}

func (e MemEventsService) GetProtoEvents(ctx context.Context) []gogoproto.Message {
	return unwrap(ctx).protoEvents[e.moduleName]
}

type eventManager struct {
	moduleName string
	ctx        *dummyCtx
}

func (e eventManager) Emit(event gogoproto.Message) error {
	e.ctx.protoEvents[e.moduleName] = append(e.ctx.protoEvents[e.moduleName], event)
	return nil
}

func (e eventManager) EmitKV(eventType string, attrs ...event.Attribute) error {
	e.ctx.events[e.moduleName] = append(e.ctx.events[e.moduleName], event.NewEvent(eventType, attrs...))
	return nil
}
