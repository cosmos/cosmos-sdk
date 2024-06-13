package indexerbase

import (
	"encoding/json"
)

// Listener is an interface that defines methods for listening to both raw and logical blockchain data.
// It is valid for any of the methods to be nil, in which case the listener will not be called for that event.
// Listeners should understand the guarantees that are provided by the source they are listening to and
// understand which methods will or will not be called. For instance, most blockchains will not do logical
// decoding of data out of the box, so the InitializeModuleSchema and OnEntityUpdate methods will not be called.
// These methods will only be called when listening logical decoding is setup.
type Listener struct {
	// Initialize is called when the listener is initialized before any other methods are called.
	// The lastBlockPersisted return value should be the last block height the listener persisted if it is
	// persisting block data, 0 if it is not interested in persisting block data, or -1 if it is
	// persisting block data but has not persisted any data yet. This check allows the indexer
	// framework to ensure that the listener has not missed blocks.
	Initialize func(InitializationData) (lastBlockPersisted int64, err error)

	// StartBlock is called at the beginning of processing a block.
	StartBlock func(uint64) error

	// OnBlockHeader is called when a block header is received.
	OnBlockHeader func(BlockHeaderData) error

	// OnTx is called when a transaction is received.
	OnTx func(TxData) error

	// OnEvent is called when an event is received.
	OnEvent func(EventData) error

	// OnKVPair is called when a key-value has been written to the store for a given module.
	OnKVPair func(moduleName string, key, value []byte, delete bool) error

	// Commit is called when state is commited, usually at the end of a block. Any
	// indexers should commit their data when this is called and return an error if
	// they are unable to commit.
	Commit func() error

	// InitializeModuleSchema should be called whenever the blockchain process starts OR whenever
	// logical decoding of a module is initiated. An indexer listening to this event
	// should ensure that they have performed whatever initialization steps (such as database
	// migrations) required to receive OnEntityUpdate events for the given module. If the
	// indexer's schema is incompatible with the module's on-chain schema, the listener should return
	// an error.
	InitializeModuleSchema func(module string, schema ModuleSchema) error

	// OnEntityUpdate is called whenever an entity is updated in the module. This is only called
	// when logical data is available. It should be assumed that the same data in raw form
	// is also passed to OnKVPair.
	OnEntityUpdate func(module string, update EntityUpdate) error

	// CommitCatchupSync is called after all existing entities for a module have been passed to
	// OnEntityUpdate during a catch-up sync which has been initiated by return -1 for lastBlock
	// in InitializeModuleSchema. The listener should commit all the data that has been received at
	// this point and also save the block number as the last block that has been processed so
	// that processing of regular block data can resume from this point in the future.
	CommitCatchupSync func(module string, block uint64) error

	// SubscribedModules is a map of modules that the listener is interested in receiving events for in OnKVPair and
	// logical decoding listeners (if these are registered). If this is left nil but listeners are registered,
	// it is assumed that the listener is interested in all modules.
	SubscribedModules map[string]bool
}

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
