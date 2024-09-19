package stf

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"slices"

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
	res, err := TypedEventToEvent(tev)
	if err != nil {
		return err
	}

	em.executionContext.events = append(em.executionContext.events, res)
	return nil
}

// EmitKV emits a key value pair event.
func (em *eventManager) EmitKV(eventType string, attrs ...event.Attribute) error {
	em.executionContext.events = append(em.executionContext.events, event.NewEvent(eventType, attrs...))
	return nil
}

// TypedEventToEvent takes typed event and converts to Event object
func TypedEventToEvent(tev transaction.Msg) (event.Event, error) {
	evtType := gogoproto.MessageName(tev)
	buf := new(bytes.Buffer)
	jm := &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}
	if err := jm.Marshal(buf, tev); err != nil {
		return event.Event{}, err
	}

	var attrMap map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &attrMap); err != nil {
		return event.Event{}, err
	}

	// sort the keys to ensure the order is always the same
	keys := slices.Sorted(maps.Keys(attrMap))
	attrs := make([]event.Attribute, 0, len(attrMap))
	for _, k := range keys {
		v := attrMap[k]
		attrs = append(attrs, event.Attribute{
			Key:   k,
			Value: string(v),
		})
	}

	return event.Event{
		Type:       evtType,
		Attributes: func() ([]event.Attribute, error) { return attrs, nil },
	}, nil
}
