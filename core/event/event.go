package event

import "encoding/json"

// Attribute is a kv-pair event attribute.
type Attribute = struct {
	Key, Value string
}

func NewAttribute(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Events represents a list of events.
type Events struct {
	Events []Event
}

func NewEvents(events ...Event) Events {
	return Events{Events: events}
}

// Event represents the data for a single event.
type Event struct {
	// BlockStage represents the stage of the block at which this event is associated.
	// If the block stage is unknown, it should be set to UnknownBlockStage.
	BlockStage BlockStage

	// TxIndex is the 1-based index of the transaction in the block to which this event is associated.
	// If TxIndex is zero, it means that we do not know the transaction index.
	// Otherwise, the index should start with 1.
	TxIndex int32

	// MsgIndex is the 1-based index of the message in the transaction to which this event is associated.
	// If MsgIndex is zero, it means that we do not know the message index.
	// Otherwise, the index should start with 1.
	MsgIndex int32

	// EventIndex is the 1-based index of the event in the message to which this event is associated.
	// If EventIndex is zero, it means that we do not know the event index.
	// Otherwise, the index should start with 1.
	EventIndex int32

	// Type is the type of the event.
	Type string

	// Data lazily returns the JSON representation of the event.
	Data ToJSON

	// Attributes lazily returns the key-value attribute representation of the event.
	Attributes ToEventAttributes
}

// BlockStage represents the stage of block processing for an event.
type BlockStage int32

const (
	// UnknownBlockStage indicates that we do not know the block stage.
	UnknownBlockStage BlockStage = iota

	// PreBlockStage indicates that the event is associated with the pre-block stage.
	PreBlockStage

	// BeginBlockStage indicates that the event is associated with the begin-block stage.
	BeginBlockStage

	// TxProcessingStage indicates that the event is associated with the transaction processing stage.
	TxProcessingStage

	// EndBlockStage indicates that the event is associated with the end-block stage.
	EndBlockStage
)

// ToJSON is a function that lazily returns the JSON representation of data.
type ToJSON = func() (json.RawMessage, error)

// ToEventAttributes is a function that lazily returns the key-value attribute representation of an event.
type ToEventAttributes = func() ([]Attribute, error)

func NewEvent(ty string, attrs ...Attribute) Event {
	return Event{Type: ty, Attributes: func() ([]Attribute, error) { return attrs, nil }}
}
