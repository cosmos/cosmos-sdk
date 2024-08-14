package appdata

// Listener is an interface that defines methods for listening to both raw and logical blockchain data.
// It is valid for any of the methods to be nil, in which case the listener will not be called for that event.
// Listeners should understand the guarantees that are provided by the source they are listening to and
// understand which methods will or will not be called. For instance, most blockchains will not do logical
// decoding of data out of the box, so the InitializeModuleData and OnObjectUpdate methods will not be called.
// These methods will only be called when listening logical decoding is setup.
type Listener struct {
	// InitializeModuleData should be called whenever the blockchain process starts OR whenever
	// logical decoding of a module is initiated. An indexer listening to this event
	// should ensure that they have performed whatever initialization steps (such as database
	// migrations) required to receive OnObjectUpdate events for the given module. If the
	// indexer's schema is incompatible with the module's on-chain schema, the listener should return
	// an error. Module names must conform to the NameFormat regular expression.
	InitializeModuleData func(ModuleInitializationData) error

	// StartBlock is called at the beginning of processing a block.
	StartBlock func(StartBlockData) error

	// OnTx is called when a transaction is received.
	OnTx func(TxData) error

	// OnEvent is called when an event is received.
	OnEvent func(EventData) error

	// OnKVPair is called when a key-value has been written to the store for a given module.
	// Module names must conform to the NameFormat regular expression.
	OnKVPair func(updates KVPairData) error

	// OnObjectUpdate is called whenever an object is updated in a module's state. This is only called
	// when logical data is available. It should be assumed that the same data in raw form
	// is also passed to OnKVPair. Module names must conform to the NameFormat regular expression.
	OnObjectUpdate func(ObjectUpdateData) error

	// Commit is called when state is committed, usually at the end of a block. Any
	// indexers should commit their data when this is called and return an error if
	// they are unable to commit. Data sources MUST call Commit when data is committed,
	// otherwise it should be assumed that indexers have not persisted their state.
	// When listener processing is pushed into background go routines using AsyncListener
	// or AsyncListenerMux, Commit will synchronize the processing of all listeners and
	// is a blocking call. Producers that do not want to block on Commit in a given block
	// can delay calling Commit until the start of the next block to give listener
	// processing time to complete.
	Commit func(CommitData) error

	// onBatch can be used internally to efficiently forward packet batches to
	// async listeners.
	onBatch func(PacketBatch) error
}
