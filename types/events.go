package types

import (
	"fmt"
	"sort"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// ----------------------------------------------------------------------------
// Event Manager
// ----------------------------------------------------------------------------

// EventManager implements a simple wrapper around a slice of Event objects that
// can be emitted from.
type EventManager struct {
	events Events
}

func NewEventManager() *EventManager {
	return &EventManager{EmptyEvents()}
}

func (em *EventManager) Events() Events { return em.events }

// EmitEvent stores a single Event object.
func (em *EventManager) EmitEvent(event Event) {
	em.events = em.events.AppendEvent(event)
}

// EmitEvents stores a series of Event objects.
func (em *EventManager) EmitEvents(events Events) {
	em.events = em.events.AppendEvents(events)
}

// ABCIEvents returns all stored Event objects as abci.Event objects.
func (em EventManager) ABCIEvents() []abci.Event {
	return em.events.ToABCIEvents()
}

// ----------------------------------------------------------------------------
// Events
// ----------------------------------------------------------------------------

type (
	// Event is a type alias for an ABCI Event
	Event abci.Event

	// Attribute defines an attribute wrapper where the key and value are
	// strings instead of raw bytes.
	Attribute struct {
		Key   string `json:"key"`
		Value string `json:"value,omitempty"`
	}

	// Events defines a slice of Event objects
	Events []Event
)

// NewEvent creates a new Event object with a given type and slice of one or more
// attributes.
func NewEvent(ty string, attrs ...Attribute) Event {
	e := Event{Type: ty}

	for _, attr := range attrs {
		e.Attributes = append(e.Attributes, NewAttribute(attr.Key, attr.Value).ToKVPair())
	}

	return e
}

// NewAttribute returns a new key/value Attribute object.
func NewAttribute(k, v string) Attribute {
	return Attribute{k, v}
}

// EmptyEvents returns an empty slice of events.
func EmptyEvents() Events {
	return make(Events, 0)
}

func (a Attribute) String() string {
	return fmt.Sprintf("%s: %s", a.Key, a.Value)
}

// ToKVPair converts an Attribute object into a Tendermint key/value pair.
func (a Attribute) ToKVPair() cmn.KVPair {
	return cmn.KVPair{Key: toBytes(a.Key), Value: toBytes(a.Value)}
}

// AppendAttributes adds one or more attributes to an Event.
func (e Event) AppendAttributes(attrs ...Attribute) Event {
	for _, attr := range attrs {
		e.Attributes = append(e.Attributes, attr.ToKVPair())
	}
	return e
}

// AppendEvent adds an Event to a slice of events.
func (e Events) AppendEvent(event Event) Events {
	return append(e, event)
}

// AppendEvents adds a slice of Event objects to an exist slice of Event objects.
func (e Events) AppendEvents(events Events) Events {
	return append(e, events...)
}

// ToABCIEvents converts a slice of Event objects to a slice of abci.Event
// objects.
func (e Events) ToABCIEvents() []abci.Event {
	res := make([]abci.Event, len(e), len(e))
	for i, ev := range e {
		res[i] = abci.Event{Type: ev.Type, Attributes: ev.Attributes}
	}

	return res
}

func toBytes(i interface{}) []byte {
	switch x := i.(type) {
	case []uint8:
		return x
	case string:
		return []byte(x)
	default:
		panic(i)
	}
}

// Common event types and attribute keys
var (
	EventTypeMessage = "message"

	AttributeKeyAction = "action"
	AttributeKeyModule = "module"
	AttributeKeySender = "sender"
	AttributeKeyAmount = "amount"
)

type (
	// StringAttribute defines en Event object wrapper where all the attributes
	// contain key/value pairs that are strings instead of raw bytes.
	StringEvent struct {
		Type       string      `json:"type,omitempty"`
		Attributes []Attribute `json:"attributes,omitempty"`
	}

	// StringAttributes defines a slice of StringEvents objects.
	StringEvents []StringEvent
)

func (se StringEvents) String() string {
	var sb strings.Builder

	for _, e := range se {
		sb.WriteString(fmt.Sprintf("\t\t- %s\n", e.Type))

		for _, attr := range e.Attributes {
			sb.WriteString(fmt.Sprintf("\t\t\t- %s\n", attr.String()))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Flatten returns a flattened version of StringEvents by grouping all attributes
// per unique event type.
func (se StringEvents) Flatten() StringEvents {
	flatEvents := make(map[string][]Attribute)

	for _, e := range se {
		flatEvents[e.Type] = append(flatEvents[e.Type], e.Attributes...)
	}

	var (
		res  StringEvents
		keys []string
	)

	for ty := range flatEvents {
		keys = append(keys, ty)
	}

	sort.Strings(keys)
	for _, ty := range keys {
		res = append(res, StringEvent{Type: ty, Attributes: flatEvents[ty]})
	}

	return res
}

// StringifyEvent converts an Event object to a StringEvent object.
func StringifyEvent(e abci.Event) StringEvent {
	res := StringEvent{Type: e.Type}

	for _, attr := range e.Attributes {
		res.Attributes = append(
			res.Attributes,
			Attribute{string(attr.Key), string(attr.Value)},
		)
	}

	return res
}

// StringifyEvents converts a slice of Event objects into a slice of StringEvent
// objects.
func StringifyEvents(events []abci.Event) StringEvents {
	var res StringEvents

	for _, e := range events {
		res = append(res, StringifyEvent(e))
	}

	return res.Flatten()
}
