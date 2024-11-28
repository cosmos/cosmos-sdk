package stf

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/cosmos/gogoproto/jsonpb"
	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/transaction"
)

func NewEventService() event.Service {
	return eventService{}
}

type eventService struct{}

// EventManager implements event.Service.
func (eventService) EventManager(ctx context.Context) event.Manager {
	exCtx, err := getExecutionCtxFromContext(ctx)
	if err != nil {
		panic(err)
	}

	return &eventManager{exCtx}
}

var _ event.Manager = (*eventManager)(nil)

type eventManager struct {
	executionContext *executionContext
}

// Emit emits an typed event that is defined in the protobuf file.
// In the future these events will be added to consensus.
func (em *eventManager) Emit(tev transaction.Msg) error {
	event := event.Event{
		Type: gogoproto.MessageName(tev),
		Data: func() (json.RawMessage, error) {
			buf := new(bytes.Buffer)
			jm := &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}
			if err := jm.Marshal(buf, tev); err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		},
	}

	em.executionContext.events = append(em.executionContext.events, event)
	return nil
}

// EmitKV emits a key value pair event.
func (em *eventManager) EmitKV(eventType string, attrs ...event.Attribute) error {
	em.executionContext.events = append(em.executionContext.events, event.NewEvent(eventType, attrs...))
	return nil
}
