package appdata

import (
	"encoding/json"

	"cosmossdk.io/schema"
)

// ModuleInitializationData represents data for related to module initialization, in particular
// the module's schema.
type ModuleInitializationData struct {
	// ModuleName is the name of the module.
	ModuleName string

	// Schema is the schema of the module.
	Schema schema.ModuleSchema
}

// StartBlockData represents the data that is passed to a listener when a block is started.
type StartBlockData struct {
	// Height is the height of the block.
	Height uint64

	// Bytes is the raw byte representation of the block header. It may be nil if the source does not provide it.
	HeaderBytes ToBytes

	// JSON is the JSON representation of the block header. It should generally be a JSON object.
	// It may be nil if the source does not provide it.
	HeaderJSON ToJSON
}

// TxData represents the raw transaction data that is passed to a listener.
type TxData struct {
	// TxIndex is the index of the transaction in the block.
	TxIndex int32

	// Bytes is the raw byte representation of the transaction.
	Bytes ToBytes

	// JSON is the JSON representation of the transaction. It should generally be a JSON object.
	JSON ToJSON
}

// EventData represents event data that is passed to a listener when events are received.
type EventData struct {
	// Events are the events that are received.
	Events []Event
}

// Event represents the data for a single event.
type Event struct {
	// TxIndex is the index of the transaction in the block to which this event is associated.
	// It should be set to a negative number if the event is not associated with a transaction.
	// Canonically -1 should be used to represent begin block processing and -2 should be used to
	// represent end block processing.
	TxIndex int32

	// MsgIndex is the index of the message in the transaction to which this event is associated.
	// If TxIndex is negative, this index could correspond to the index of the message in
	// begin or end block processing if such indexes exist, or it can be set to zero.
	MsgIndex int32

	// EventIndex is the index of the event in the message to which this event is associated.
	EventIndex int32

	// Type is the type of the event.
	Type string

	// Data lazily returns the JSON representation of the event.
	Data ToJSON

	// Attributes lazily returns the key-value attribute representation of the event.
	Attributes ToEventAttributes
}

type EventAttribute = struct {
	Key, Value string
}

// ToBytes is a function that lazily returns the raw byte representation of data.
type ToBytes = func() ([]byte, error)

// ToJSON is a function that lazily returns the JSON representation of data.
type ToJSON = func() (json.RawMessage, error)

// ToEventAttributes is a function that lazily returns the key-value attribute representation of an event.
type ToEventAttributes = func() ([]EventAttribute, error)

// KVPairData represents a batch of key-value pair data that is passed to a listener.
type KVPairData struct {
	Updates []ActorKVPairUpdate
}

// ActorKVPairUpdate represents a key-value pair update for a specific module or account.
type ActorKVPairUpdate = struct {
	// Actor is the byte representation of the module or account that is updating the key-value pair.
	Actor []byte

	// StateChanges are key-value pair updates.
	StateChanges []schema.KVPairUpdate
}

// ObjectUpdateData represents object update data that is passed to a listener.
type ObjectUpdateData struct {
	// ModuleName is the name of the module that the update corresponds to.
	ModuleName string

	// Updates are the object updates.
	Updates []schema.ObjectUpdate
}

// CommitData represents commit data. It is empty for now, but fields could be added later.
type CommitData struct{}
