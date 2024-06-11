package indexerbase

import (
	"encoding/json"
)

// Listener is an interface that defines methods for listening to both raw and logical blockchain data.
// It is valid for any of the methods to be nil, in which case the listener will not be called for that event.
// Listeners should understand the guarantees that are provided by the source they are listening to and
// understand which methods will or will not be called. For instance, most blockchains will not do logical
// decoding of data out of the box, so the EnsureLogicalSetup and OnEntityUpdate methods will not be called.
// These methods will only be called when listening logical decoding is setup.
type Listener struct {
	// StartBlock is called at the beginning of processing a block.
	StartBlock func(uint64) error

	// OnBlockHeader is called when a block header is received.
	OnBlockHeader func(BlockHeaderData) error

	// OnTx is called when a transaction is received.
	OnTx func(TxData) error

	// OnEvent is called when an event is received.
	OnEvent func(EventData) error

	// OnKVPair is called when a key-value has been written to the store for a given module.
	OnKVPair func(module string, key, value []byte, delete bool) error

	// Commit is called when state is commited, usually at the end of a block. Any
	// indexers should commit their data when this is called and return an error if
	// they are unable to commit.
	Commit func() error

	// EnsureLogicalSetup should be called whenever the blockchain process starts OR whenever
	// logical decoding of a module is initiated. An indexer listening to this event
	// should ensure that they have performed whatever initialization steps (such as database
	// migrations) required to receive OnEntityUpdate events for the given module. If the
	// schema is incompatible with the existing schema, the listener should return an error.
	// If the listener is persisting state for the module, it should return the last block
	// that was saved for the module so that the framework can determine whether it is safe
	// to resume indexing from the current height or whether there is a gap (usually an error).
	// If the listener does not persist any state for the module, it should return 0 for lastBlock
	// and nil for error.
	// If the listener has initialized properly and would like to persist state for the module,
	// but does not have any persisted state yet, it should return -1 for lastBlock and nil for error.
	// In this case, the framework will perform a "catch-up sync" calling OnEntityUpdate for every
	// entity already in the module followed by CommitCatchupSync before processing new block data.
	EnsureLogicalSetup func(module string, schema ModuleSchema) (lastBlock int64, err error)

	// OnEntityUpdate is called whenever an entity is updated in the module. This is only called
	// when logical data is available. It should be assumed that the same data in raw form
	// is also passed to OnKVPair.
	OnEntityUpdate func(module string, update EntityUpdate) error

	// CommitCatchupSync is called after all existing entities for a module have been passed to
	// OnEntityUpdate during a catch-up sync which has been initiated by return -1 for lastBlock
	// in EnsureLogicalSetup. The listener should commit all the data that has been received at
	// this point and also save the block number as the last block that has been processed so
	// that processing of regular block data can resume from this point in the future.
	CommitCatchupSync func(module string, block uint64) error
}

// BlockHeaderData represents the raw block header data that is passed to a listener.
type BlockHeaderData struct {
	// Height is the height of the block.
	Height uint64

	// Bytes is the raw byte representation of the block header.
	Bytes ToBytes

	// JSON is the JSON representation of the block header. It should generally be a JSON object.
	JSON ToJSON
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
