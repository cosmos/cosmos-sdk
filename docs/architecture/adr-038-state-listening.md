# ADR 038: KVStore state listening

## Changelog

- 11/23/2020: Initial draft

## Status

Proposed

## Abstract

This ADR defines a set of changes to enable listening to state changes of individual KVStores and exposing these data to consumers.

## Context

Currently, KVStore data can be remotely accessed through [Queries](https://github.com/cosmos/cosmos-sdk/blob/master/docs/building-modules/messages-and-queries.md#queries)
which proceed either through Tendermint and the ABCI, or through the gRPC server.
In addition to these request/response queries, it would be beneficial to have a means of listening to state changes as they occur in real time.

## Decision

We will modify the `MultiStore` interface and its concrete (`rootmulti` and `cachemulti`) implementations and introduce a new `listenkv.Store` to allow listening to state changes in underlying KVStores.
We will also introduce the tooling for writing these state changes out to files and configuring this service.

### Listening interface

In a new file, `store/types/listening.go`, we will create a `WriteListener` interface for streaming out state changes from a KVStore.

```go
// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the the same WriteListener across separate KVStores
	// delete bool indicates if it was a delete; true: delete, false: set
    OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool) error
}
```

### Listener type

We will create a concrete implementation of the `WriteListener` interface in `store/types/listening.go`, that writes out protobuf
encoded KV pairs to an underlying `io.Writer`.

This will include defining a simple protobuf type for the KV pairs. In addition to the key and value fields this message
will include the StoreKey for the originating KVStore so that we can write out from separate KVStores to the same stream/file
and determine the source of each KV pair.

```protobuf
message StoreKVPair {
  optional string store_key = 1; // the store key for the KVStore this pair originates from
  required bool set = 2; // true indicates a set operation, false indicates a delete operation
  required bytes key = 3;
  required bytes value = 4;
}
```

```go
// StoreKVPairWriteListener is used to configure listening to a KVStore by writing out length-prefixed
// protobuf encoded StoreKVPairs to an underlying io.Writer
type StoreKVPairWriteListener struct {
	writer io.Writer
	marshaller codec.BinaryCodec
}

// NewStoreKVPairWriteListener wraps creates a StoreKVPairWriteListener with a provdied io.Writer and codec.BinaryCodec
func NewStoreKVPairWriteListener(w io.Writer, m codec.BinaryCodec) *StoreKVPairWriteListener {
	return &StoreKVPairWriteListener{
		writer: w,
		marshaller: m,
	}
}

// OnWrite satisfies the WriteListener interface by writing length-prefixed protobuf encoded StoreKVPairs
func (wl *StoreKVPairWriteListener) OnWrite(storeKey types.StoreKey, key []byte, value []byte, delete bool) error error {
	kvPair := new(types.StoreKVPair)
	kvPair.StoreKey = storeKey.Name()
	kvPair.Delete = Delete
	kvPair.Key = key
	kvPair.Value = value
	by, err := wl.marshaller.MarshalBinaryLengthPrefixed(kvPair)
	if err != nil {
                return err
	}
        if _, err := wl.writer.Write(by); err != nil {
        	return err
        }
        return nil
}
```

### ListenKVStore

We will create a new `Store` type `listenkv.Store` that the `MultiStore` wraps around a `KVStore` to enable state listening.
We can configure the `Store` with a set of `WriteListener`s which stream the output to specific destinations.

```go
// Store implements the KVStore interface with listening enabled.
// Operations are traced on each core KVStore call and written to any of the
// underlying listeners with the proper key and operation permissions
type Store struct {
	parent    types.KVStore
	listeners []types.WriteListener
	parentStoreKey types.StoreKey
}

// NewStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewStore(parent types.KVStore, psk types.StoreKey, listeners []types.WriteListener) *Store {
	return &Store{parent: parent, listeners: listeners, parentStoreKey: psk}
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (s *Store) Set(key []byte, value []byte) {
	types.AssertValidKey(key)
	s.parent.Set(key, value)
	s.onWrite(false, key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (s *Store) Delete(key []byte) {
	s.parent.Delete(key)
	s.onWrite(true, key, nil)
}

// onWrite writes a KVStore operation to all of the WriteListeners
func (s *Store) onWrite(delete bool, key, value []byte) {
	for _, l := range s.listeners {
		if err := l.OnWrite(s.parentStoreKey, key, value, delete); err != nil {
                    // log error
                }
	}
}
```

### MultiStore interface updates

We will update the `MultiStore` interface to allow us to wrap a set of listeners around a specific `KVStore`.
Additionally, we will update the `CacheWrap` and `CacheWrapper` interfaces to enable listening in the caching layer.

```go
type MultiStore interface {
	...

	// ListeningEnabled returns if listening is enabled for the KVStore belonging the provided StoreKey
	ListeningEnabled(key StoreKey) bool

	// AddListeners adds WriteListeners for the KVStore belonging to the provided StoreKey
	// It appends the listeners to a current set, if one already exists
	AddListeners(key StoreKey, listeners []WriteListener)
}
```

```go
type CacheWrap interface {
	...

	// CacheWrapWithListeners recursively wraps again with listening enabled
	CacheWrapWithListeners(storeKey types.StoreKey, listeners []WriteListener) CacheWrap
}

type CacheWrapper interface {
	...

	// CacheWrapWithListeners recursively wraps again with listening enabled
	CacheWrapWithListeners(storeKey types.StoreKey, listeners []WriteListener) CacheWrap
}
```

### MultiStore implementation updates

We will modify all of the `Store` and `MultiStore` implementations to satisfy these new interfaces, and adjust the `rootmulti` `GetKVStore` method
to wrap the returned `KVStore` with a `listenkv.Store` if listening is turned on for that `Store`.

```go
func (rs *Store) GetKVStore(key types.StoreKey) types.KVStore {
	store := rs.stores[key].(types.KVStore)

	if rs.TracingEnabled() {
		store = tracekv.NewStore(store, rs.traceWriter, rs.traceContext)
	}
	if rs.ListeningEnabled(key) {
		store = listenkv.NewStore(key, store, rs.listeners[key])
	}

	return store
}
```

We will also adjust the `cachemulti` constructor methods and the `rootmulti` `CacheMultiStore` method to forward the listeners
to and enable listening in the cache layer.

```go
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}
	return cachemulti.NewStore(rs.db, stores, rs.keysByName, rs.traceWriter, rs.traceContext, rs.listeners)
}
```

### Exposing the data

We will introduce a new `StreamingService` interface for exposing `WriteListener` data streams to external consumers.
In addition to streaming state changes as `StoreKVPair`s, the interface satisfies an `ABCIListener` interface that plugs into the BaseApp
and relays ABCI requests and responses so that the service can group the state changes with the ABCI requests that affected them and the ABCI responses they affected.

```go
// ABCIListener interface used to hook into the ABCI message processing of the BaseApp
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages 
	ListenBeginBlock(ctx types.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error 
	// ListenEndBlock updates the steaming service with the latest EndBlock messages 
	ListenEndBlock(ctx types.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error 
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages 
	ListenDeliverTx(ctx types.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error
}

// StreamingService interface for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService interface {
	// Stream is the streaming service loop, awaits kv pairs and writes them to some destination stream or file 
	Stream(wg *sync.WaitGroup) error
	// Listeners returns the streaming service's listeners for the BaseApp to register 
	Listeners() map[types.StoreKey][]store.WriteListener 
	// ABCIListener interface for hooking into the ABCI messages from inside the BaseApp 
	ABCIListener 
	// Closer interface 
	io.Closer
}
```

#### Writing state changes to files

We will introduce an implementation of `StreamingService` which writes state changes out to files as length-prefixed protobuf encoded `StoreKVPair`s.
This service uses the same `StoreKVPairWriteListener` for every KVStore, writing all the KV pairs from every KVStore
out to the same files, relying on the `StoreKey` field in the `StoreKVPair` protobuf message to later distinguish the source for each pair.

Writing to a file is the simplest approach for streaming the data out to consumers.
This approach also provides the advantages of being persistent and durable, and the files can be read directly,
or an auxiliary streaming services can read from the files and serve the data over a remote interface.

##### Encoding

For each pair of `BeginBlock` requests and responses, a file is created and named `block-{N}-begin`, where N is the block number.
At the head of this file the length-prefixed protobuf encoded `BeginBlock` request is written.
At the tail of this file the length-prefixed protobuf encoded `BeginBlock` response is written.
In between these two encoded messages, the state changes that occurred due to the `BeginBlock` request are written chronologically as
a series of length-prefixed protobuf encoded `StoreKVPair`s representing `Set` and `Delete` operations within the KVStores the service
is configured to listen to.

For each pair of `DeliverTx` requests and responses, a file is created and named `block-{N}-tx-{M}` where N is the block number and M
is the tx number in the block (i.e. 0, 1, 2...).
At the head of this file the length-prefixed protobuf encoded `DeliverTx` request is written.
At the tail of this file the length-prefixed protobuf encoded `DeliverTx` response is written.
In between these two encoded messages, the state changes that occurred due to the `DeliverTx` request are written chronologically as
a series of length-prefixed protobuf encoded `StoreKVPair`s representing `Set` and `Delete` operations within the KVStores the service
is configured to listen to.

For each pair of `EndBlock` requests and responses, a file is created and named `block-{N}-end`, where N is the block number.
At the head of this file the length-prefixed protobuf encoded `EndBlock` request is written.
At the tail of this file the length-prefixed protobuf encoded `EndBlock` response is written.
In between these two encoded messages, the state changes that occurred due to the `EndBlock` request are written chronologically as
a series of length-prefixed protobuf encoded `StoreKVPair`s representing `Set` and `Delete` operations within the KVStores the service
is configured to listen to.

##### Decoding

To decode the files written in the above format we read all the bytes from a given file into memory and segment them into proto
messages based on the length-prefixing of each message. Once segmented, it is known that the first message is the ABCI request,
the last message is the ABCI response, and that every message in between is a `StoreKVPair`. This enables us to decode each segment into
the appropriate message type.

The type of ABCI req/res, the block height, and the transaction index (where relevant) is known
from the file name, and the KVStore each `StoreKVPair` originates from is known since the `StoreKey` is included as a field in the proto message.

##### Implementation example

```go
// FileStreamingService is a concrete implementation of StreamingService that writes state changes out to a file
type FileStreamingService struct {
	listeners map[sdk.StoreKey][]storeTypes.WriteListener // the listeners that will be initialized with BaseApp
	srcChan <-chan []byte // the channel that all of the WriteListeners write their data out to
	filePrefix string // optional prefix for each of the generated files
	writeDir string // directory to write files into
	dstFile *os.File // the current write output file
	marshaller codec.BinaryCodec // marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	stateCache [][]byte // cache the protobuf binary encoded StoreKVPairs in the order they are received
}
```

This streaming service uses a single instance of a simple intermediate `io.Writer` as the underlying `io.Writer` for its single `StoreKVPairWriteListener`,
It collects KV pairs from every KVStore synchronously off of the same channel, caching them in the order they are received, and then writing
them out to a file generated in response to an ABCI message hook. Files are named as outlined above, with optional prefixes to avoid potential naming collisions
across separate instances.

```go
// intermediateWriter is used so that we do not need to update the underlying io.Writer inside the StoreKVPairWriteListener
// everytime we begin writing to a new file
type intermediateWriter struct {
	outChan chan <-[]byte
}

// NewIntermediateWriter create an instance of an intermediateWriter that sends to the provided channel
func NewIntermediateWriter(outChan chan <-[]byte) *intermediateWriter {
	return &intermediateWriter{
		outChan: outChan,
	}
}

// Write satisfies io.Writer
func (iw *intermediateWriter) Write(b []byte) (int, error) {
	iw.outChan <- b
	return len(b), nil
}

// NewFileStreamingService creates a new FileStreamingService for the provided writeDir, (optional) filePrefix, and storeKeys
func NewFileStreamingService(writeDir, filePrefix string, storeKeys []sdk.StoreKey, m codec.BinaryCodec) (*FileStreamingService, error) {
	listenChan := make(chan []byte, 0)
	iw := NewIntermediateWriter(listenChan)
	listener := listen.NewStoreKVPairWriteListener(iw, m)
	listners := make(map[sdk.StoreKey][]storeTypes.WriteListener, len(storeKeys))
	// in this case, we are using the same listener for each Store
	for _, key := range storeKeys {
		listeners[key] = listener
	}
	// check that the writeDir exists and is writeable so that we can catch the error here at initialization if it is not
	// we don't open a dstFile until we receive our first ABCI message
	if err := fileutil.IsDirWriteable(writeDir); err != nil {
		return nil, err
	}
	return &FileStreamingService{
		listeners: listeners,
		srcChan: listenChan,
		filePrefix: filePrefix,
		writeDir: writeDir,
		marshaller: m,
		stateCache: make([][]byte, 0),
	}, nil
}

// Listeners returns the StreamingService's underlying WriteListeners, use for registering them with the BaseApp
func (fss *FileStreamingService) Listeners() map[sdk.StoreKey][]storeTypes.WriteListener {
	return fss.listeners
}

func (fss *FileStreamingService) ListenBeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	// NOTE: this could either be done synchronously or asynchronously
	// create a new file with the req info according to naming schema
	// write req to file
	// write all state changes cached for this stage to file
	// reset cache
	// write res to file
	// close file
}

func (fss *FileStreamingService) ListenEndBlock(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) {
	// NOTE: this could either be done synchronously or asynchronously
	// create a new file with the req info according to naming schema
	// write req to file
	// write all state changes cached for this stage to file
	// reset cache
	// write res to file
	// close file
}

func (fss *FileStreamingService) ListenDeliverTx(ctx sdk.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) {
	// NOTE: this could either be done synchronously or asynchronously
	// create a new file with the req info according to naming schema
	// NOTE: if the tx failed, handle accordingly
	// write req to file
	// write all state changes cached for this stage to file
	// reset cache
	// write res to file
	// close file
}

// Stream spins up a goroutine select loop which awaits length-prefixed binary encoded KV pairs and caches them in the order they were received
func (fss *FileStreamingService) Stream(wg *sync.WaitGroup, quitChan <-chan struct{}) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quitChan:
				return
                        case by := <-fss.srcChan:
                        	fss.stateCache = append(fss.stateCache, by)
			}
		}
	}()
}
```

#### Auxiliary streaming service

We will create a separate standalone process that reads and internally queues the state as it is written out to these files
and serves the data over a gRPC API. This API will allow filtering of requested data, e.g. by block number, block/tx hash, ABCI message type,
whether a DeliverTx message failed or succeeded, etc. In addition to unary RPC endpoints this service will expose `stream` RPC endpoints for realtime subscriptions.

#### File pruning

Without pruning the number of files can grow indefinitely, this may need to be managed by
the developer in an application or even module-specific manner (e.g. log rotation).
The file naming schema facilitates pruning by block number and/or ABCI message.
The gRPC auxiliary streaming service introduced above will include an option to remove the files as it consumes their data.

### Configuration

We will provide detailed documentation on how to configure a `FileStreamingService` from within an app's `AppCreator`,
using the provided `AppOptions` and TOML configuration fields.

#### BaseApp registration

We will add a new method to the `BaseApp` to enable the registration of `StreamingService`s:

```go
// SetStreamingService is used to register a streaming service with the BaseApp
func (app *BaseApp) SetStreamingService(s StreamingService) {
	// set the listeners for each StoreKey
	for key, lis := range s.Listeners() {
		app.cms.AddListeners(key, lis)
	}
	// register the streaming service hooks within the BaseApp
	// BaseApp will pass BeginBlock, DeliverTx, and EndBlock requests and responses to the streaming services to update their ABCI context using these hooks
	app.hooks = append(app.hooks, s)
}
```

We will also modify the `BeginBlock`, `EndBlock`, and `DeliverTx` methods to pass ABCI requests and responses to any streaming service hooks registered
with the `BaseApp`.

```go
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {

	...

	// Call the streaming service hooks with the BeginBlock messages
	for _, hook := range app.hooks {
		hook.ListenBeginBlock(app.deliverState.ctx, req, res)
	}

	return res
}
```

```go
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {

	...

	// Call the streaming service hooks with the EndBlock messages
	for _, hook := range app.hooks {
		hook.ListenEndBlock(app.deliverState.ctx, req, res)
	}

	return res
}
```

```go
func (app *BaseApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {

	...

	gInfo, result, err := app.runTx(runTxModeDeliver, req.Tx)
	if err != nil {
		resultStr = "failed"
		res := sdkerrors.ResponseDeliverTx(err, gInfo.GasWanted, gInfo.GasUsed, app.trace)
		// If we throw and error, be sure to still call the streaming service's hook
		for _, hook := range app.hooks {
			hook.ListenDeliverTx(app.deliverState.ctx, req, res)
		}
		return res
	}

	res := abci.ResponseDeliverTx{
		GasWanted: int64(gInfo.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(gInfo.GasUsed),   // TODO: Should type accept unsigned ints?
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}

	// Call the streaming service hooks with the DeliverTx messages
	for _, hook := range app.hooks {
		hook.ListenDeliverTx(app.deliverState.ctx, req, res)
	}

	return res
}
```

#### TOML Configuration

We will provide standard TOML configuration options for configuring a `FileStreamingService` for specific `Store`s.
Note: the actual namespace is TBD.

```toml
[store]
    streamers = [ # if len(streamers) > 0 we are streaming
        "file",
    ]

[streamers]
    [streamers.file]
        keys = ["list", "of", "store", "keys", "we", "want", "to", "expose", "for", "this", "streaming", "service"]
        write_dir = "path to the write directory"
        prefix = "optional prefix to prepend to the generated file names"
```

We will also provide a mapping of the TOML `store.streamers` "file" configuration option to a helper functions for constructing the specified
streaming service. In the future, as other streaming services are added, their constructors will be added here as well.

Each configured streamer will receive the

```go
// ServiceConstructor is used to construct a streaming service
type ServiceConstructor func(opts serverTypes.AppOptions, keys []sdk.StoreKey, marshaller codec.BinaryMarshaler) (sdk.StreamingService, error)

// ServiceType enum for specifying the type of StreamingService
type ServiceType int

const (
  Unknown ServiceType = iota
  File
  // add more in the future
)

// NewStreamingServiceType returns the streaming.ServiceType corresponding to the provided name
func NewStreamingServiceType(name string) ServiceType {
  switch strings.ToLower(name) {
  case "file", "f":
    return File
  default:
    return Unknown
  }
}

// String returns the string name of a streaming.ServiceType
func (sst ServiceType) String() string {
  switch sst {
  case File:
    return "file"
  default:
    return ""
  }
}

// ServiceConstructorLookupTable is a mapping of streaming.ServiceTypes to streaming.ServiceConstructors
var ServiceConstructorLookupTable = map[ServiceType]ServiceConstructor{
  File: NewFileStreamingService,
}

// ServiceTypeFromString returns the streaming.ServiceConstructor corresponding to the provided name
func ServiceTypeFromString(name string) (ServiceConstructor, error) {
  ssType := NewStreamingServiceType(name)
  if ssType == Unknown {
    return nil, fmt.Errorf("unrecognized streaming service name %s", name)
  }
  if constructor, ok := ServiceConstructorLookupTable[ssType]; ok {
    return constructor, nil
  }
  return nil, fmt.Errorf("streaming service constructor of type %s not found", ssType.String())
}

// NewFileStreamingService is the streaming.ServiceConstructor function for creating a FileStreamingService
func NewFileStreamingService(opts serverTypes.AppOptions, keys []sdk.StoreKey, marshaller codec.BinaryMarshaler) (sdk.StreamingService, error) {
  filePrefix := cast.ToString(opts.Get("streamers.file.prefix"))
  fileDir := cast.ToString(opts.Get("streamers.file.write_dir"))
  return file.NewStreamingService(fileDir, filePrefix, keys, marshaller)
}
```

#### Example configuration

As a demonstration, we will implement the state watching features as part of SimApp.
For example, the below is a very rudimentary integration of the state listening features into the SimApp `AppCreator` function:

```go
func NewSimApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig simappparams.EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {

	...

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
	)

	// configure state listening capabilities using AppOptions
	listeners := cast.ToStringSlice(appOpts.Get("store.streamers"))
	for _, listenerName := range listeners {
		// get the store keys allowed to be exposed for this streaming service 
		exposeKeyStrs := cast.ToStringSlice(appOpts.Get(fmt.Sprintf("streamers.%s.keys", streamerName)))
		var exposeStoreKeys []sdk.StoreKey
		if exposeAll(exposeKeyStrs) { // if list contains `*`, expose all StoreKeys 
			exposeStoreKeys = make([]sdk.StoreKey, 0, len(keys))
			for _, storeKey := range keys {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		} else {
			exposeStoreKeys = make([]sdk.StoreKey, 0, len(exposeKeyStrs))
			for _, keyStr := range exposeKeyStrs {
				if storeKey, ok := keys[keyStr]; ok {
					exposeStoreKeys = append(exposeStoreKeys, storeKey)
				}
			}
		}
		if len(exposeStoreKeys) == 0 { // short circuit if we are not exposing anything 
			continue
		}
		// get the constructor for this listener name
		constructor, err := baseapp.NewStreamingServiceConstructor(listenerName)
		if err != nil {
			tmos.Exit(err.Error()) // or continue?
		}
		// generate the streaming service using the constructor, appOptions, and the StoreKeys we want to expose
		streamingService, err := constructor(appOpts, exposeStoreKeys, appCodec)
		if err != nil {
			tmos.Exit(err.Error())
		}
		// register the streaming service with the BaseApp
		bApp.RegisterStreamingService(streamingService)
		// waitgroup and quit channel for optional shutdown coordination of the streaming service
		wg := new(sync.WaitGroup)
		quitChan := make(chan struct{}))
		// kick off the background streaming service loop
		streamingService.Stream(wg, quitChan) // maybe this should be done from inside BaseApp instead?
	}

	...

	return app
}
```

## Consequences

These changes will provide a means of subscribing to KVStore state changes in real time.

### Backwards Compatibility

- This ADR changes the `MultiStore`, `CacheWrap`, and `CacheWrapper` interfaces, implementations supporting the previous version of these interfaces will not support the new ones

### Positive

- Ability to listen to KVStore state changes in real time and expose these events to external consumers

### Negative

- Changes `MultiStore`, `CacheWrap`, and `CacheWrapper` interfaces

### Neutral

- Introduces additional- but optional- complexity to configuring and running a cosmos application
- If an application developer opts to use these features to expose data, they need to be aware of the ramifications/risks of that data exposure as it pertains to the specifics of their application
