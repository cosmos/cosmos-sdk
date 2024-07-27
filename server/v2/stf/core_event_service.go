package stf

import (
	"context"
	"encoding/json"
	"slices"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"golang.org/x/exp/maps"

	"cosmossdk.io/core/event"
)

func NewEventService() event.Service {
	return eventService{}
}

type eventService struct{}

// EventManager implements event.Service.
func (eventService) EventManager(ctx context.Context) event.Manager {
	return &eventManager{ctx.(*executionContext)}
}

var _ event.Manager = (*eventManager)(nil)

type eventManager struct {
	executionContext *executionContext
}

// Emit emits an typed event that is defined in the protobuf file.
// In the future these events will be added to consensus.
func (em *eventManager) Emit(tev gogoproto.Message) error {
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

// EmitNonConsensus emits an typed event that is defined in the protobuf file.
// These events will not be added to consensus.
func (em *eventManager) EmitNonConsensus(event gogoproto.Message) error {
	return em.Emit(event)
}

// TypedEventToEvent takes typed event and converts to Event object
func TypedEventToEvent(tev gogoproto.Message) (event.Event, error) {
	evtType := gogoproto.MessageName(tev)
	evtJSON, err := gogoproto.Marshal(tev)
	if err != nil {
		return event.Event{}, err
	}

	var attrMap map[string]json.RawMessage
	err = json.Unmarshal(evtJSON, &attrMap)
	if err != nil {
		return event.Event{}, err
	}

	// sort the keys to ensure the order is always the same
	keys := maps.Keys(attrMap)
	slices.Sort(keys)

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
		Attributes: attrs,
	}, nil
}
