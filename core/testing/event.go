package coretesting

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
)

var _ event.Service = &TestEventService{}

type TestEventService struct {
	moduleName string
}

// NewTestEventService attaches an event service to the context.
// Adding an existing module will reset the events.
func NewTestEventService(ctx context.Context, moduleName string) TestEventService {
	unwrap(ctx).events[moduleName] = nil
	unwrap(ctx).protoEvents[moduleName] = nil
	return TestEventService{moduleName: moduleName}
}

func (e TestEventService) EventManager(ctx context.Context) event.Manager {
	return eventManager{moduleName: e.moduleName, ctx: unwrap(ctx)}
}

func (e TestEventService) GetEvents(ctx context.Context) []event.Event {
	return unwrap(ctx).events[e.moduleName]
}

func (e TestEventService) GetProtoEvents(ctx context.Context) []transaction.Msg {
	return unwrap(ctx).protoEvents[e.moduleName]
}

type eventManager struct {
	moduleName string
	ctx        *dummyCtx
}

func (e eventManager) Emit(event transaction.Msg) error {
	e.ctx.protoEvents[e.moduleName] = append(e.ctx.protoEvents[e.moduleName], event)
	return nil
}

func (e eventManager) EmitKV(eventType string, attrs ...event.Attribute) error {
	e.ctx.events[e.moduleName] = append(e.ctx.events[e.moduleName], event.NewEvent(eventType, attrs...))
	return nil
}
