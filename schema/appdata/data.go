package appdata

import (
	"encoding/json"

	"cosmossdk.io/schema"
)

// InitializationData represents initialization data that is passed to a listener.
type InitializationData struct {
	// HasEventAlignedWrites indicates that the blockchain data source will emit KV-pair events
	// in an order aligned with transaction, message and event callbacks. If this is true
	// then indexers can assume that KV-pair data is associated with these specific transactions, messages
	// and events. This may be useful for indexers which store a log of all operations (such as immutable
	// or version controlled databases) so that the history log can include fine grain correlation between
	// state updates and transactions, messages and events. If this value is false, then indexers should
	// assume that KV-pair data occurs out of order with respect to transaction, message and event callbacks -
	// the only safe assumption being that KV-pair data is associated with the block in which it was emitted.
	HasEventAlignedWrites bool
}

type ModuleInitializationData struct {
	ModuleName string
	Schema     schema.ModuleSchema
}

type StartBlockData struct {
	// Height is the height of the block.
	Height uint64

	// Bytes is the raw byte representation of the block header.
	HeaderBytes ToBytes

	// JSON is the JSON representation of the block header. It should generally be a JSON object.
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

// EventData represents event data that is passed to a listener.
type EventData struct {
	// TxIndex is the index of the transaction in the block to which this event is associated.
	// It should be set to a negative number if the event is not associated with a transaction.
	// Canonically -1 should be used to represent begin block processing and -2 should be used to
	// represent end block processing.
	TxIndex int32

	// MsgIndex is the index of the message in the transaction to which this event is associated.
	// If TxIndex is negative, this index could correspond to the index of the message in
	// begin or end block processing if such indexes exist, or it can be set to zero.
	MsgIndex uint32

	// EventIndex is the index of the event in the message to which this event is associated.
	EventIndex uint32

	// Type is the type of the event.
	Type string

	// Data is the JSON representation of the event data. It should generally be a JSON object.
	Data ToJSON
}

// ToBytes is a function that lazily returns the raw byte representation of data.
type ToBytes = func() ([]byte, error)

// ToJSON is a function that lazily returns the JSON representation of data.
type ToJSON = func() (json.RawMessage, error)

type KVPairData struct {
	Update []schema.KVPairUpdate
}

type ModuleKVPairUpdate struct {
	// ModuleName is the name of the module that the key-value pair belongs to.
	ModuleName string

	Update schema.KVPairUpdate
}

// ObjectUpdateData represents object update data that is passed to a listener.
type ObjectUpdateData struct {
	// ModuleName is the name of the module that the update corresponds to.
	ModuleName string

	// Updates are the object updates.
	Updates []schema.ObjectUpdate
}

type CommitData struct{}
