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

In a new file- `store/types/listening.go`- we will create a `WriteListener` interface for streaming out state changes from a KVStore.

```go
// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	//storeKey indicates the source KVStore, to facilitate using the the same WriteListener across separate KVStores
	OnWrite(storeKey types.StoreKey, key []byte, value []byte)
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
  optional string store_key = 1;
  required bytes key = 2;
  required bytes value = 3;
}
```

```go
// StoreKVPairWriteListener is used to configure listening to a KVStore by writing out length-prefixed
// protobuf encoded StoreKVPairs to an underlying io.Writer
type StoreKVPairWriteListener struct {
	writer io.Writer
	marshaler codec.BinaryMarshaler
}

// NewStoreKVPairWriteListener wraps creates a StoreKVPairWriteListener with a provdied io.Writer and codec.BinaryMarshaler
func NewStoreKVPairWriteListener(w io.Writer, m codec.BinaryMarshaler) *StoreKVPairWriteListener {
	return &StoreKVPairWriteListener{
		writer: w,
		marshaler: m,
	}
}

// OnWrite satisfies the WriteListener interface by writing length-prefixed protobuf encoded StoreKVPairs
func (wl *StoreKVPairWriteListener) OnWrite(storeKey types.StoreKey, key []byte, value []byte) {
	kvPair := new(types.StoreKVPair)
	kvPair.StoreKey = storeKey.Name()
	kvPair.Key = key
	kvPair.Value = value
	if by, err := wl.marshaler.MarshalBinaryLengthPrefixed(kvPair); err == nil {
		wl.writer.Write(by)
	}
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
	s.onWrite(key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (s *Store) Delete(key []byte) {
	s.parent.Delete(key)
	s.onWrite(key, nil)
}

// onWrite writes a KVStore operation to all of the WriteListeners
func (s *Store) onWrite(key, value []byte) {
	for _, l := range s.listeners {
		l.OnWrite(s.parentStoreKey, key, value)
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

	// SetListeners sets the WriteListeners for the KVStore belonging to the provided StoreKey
	// It appends the listeners to a current set, if one already exists
	SetListeners(key StoreKey, listeners []WriteListener)
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

```go
// StreamingService interface for registering WriteListeners with the BaseApp and updating the service with the ABCI context
type StreamingService interface {
	Stream(wg *sync.WaitGroup, quitChan <-chan struct{}) // streaming service loop, awaits kv pairs and writes them to some destination stream or file
	Listeners() map[sdk.StoreKey][]storeTypes.WriteListener // returns the streaming service's listeners for the BaseApp to register
	BeginBlockReq(req abci.RequestBeginBlock) // update the streaming service with the latest RequestBeginBlock message
	BeginBlockRes(res abci.ResponseBeginBlock) // update the steaming service with the latest ResponseBeginBlock message
	EndBlockReq(req abci.RequestEndBlock) // update the steaming service with the latest RequestEndBlock message
	EndBlockRes(res abci.ResponseEndBlock) // update the steaming service with the latest ResponseEndBlock message
	DeliverTxReq(req abci.RequestDeliverTx) // update the steaming service with the latest RequestDeliverTx message
	DeliverTxRes(res abci.ResponseDeliverTx) // update the steaming service with the latest ResponseDeliverTx message
}
```

#### Writing state changes to files

We will introduce an implementation of `StreamingService` which writes state changes out to files as length-prefixed protobuf encoded `StoreKVPair`s.
This service uses the same `StoreKVPairWriteListener` for every KVStore, writing all the KV pairs from every KVStore
out to the same files, relying on the `StoreKey` field in the `StoreKVPair` protobuf message to later distinguish the source KVStore for each pair.

The file naming schema is as such:
* After every `BeginBlock` request a new file is created with the name `block-{N}-begin`, where N is the block number. All
subsequent state changes are written out to this file until the first `DeliverTx` request is received. At the head of these files,
  the length-prefixed protobuf encoded `BeginBlock` request is written, and the response is written at the tail.
* After every `DeliverTx` request a new file is created with the name `block-{N}-tx-{M}` where N is the block number and M
is the tx number in the block (i.e. 0, 1, 2...). All subsequent state changes are written out to this file until the next
`DeliverTx` request is received or an `EndBlock` request is received. At the head of these files, the length-prefixed protobuf
  encoded `DeliverTx` request is written, and the response is written at the tail.
* After every `EndBlock` request a new file is created with the name `block-{N}-end`, where N is the block number. All
subsequent state changes are written out to this file until the next `BeginBlock` request is received. At the head of these files,
  the length-prefixed protobuf encoded `EndBlock` request is written, and the response is written at the tail.

```go
// FileStreamingService is a concrete implementation of StreamingService that writes state changes out to a file
type FileStreamingService struct {
	listeners map[sdk.StoreKey][]storeTypes.WriteListener // the listeners that will be initialized with BaseApp
	srcChan <-chan []byte // the channel that all of the WriteListeners write their out to
	filePrefix string // optional prefix for each of the generated files
	writeDir string // directory to write files into
	dstFile *os.File // the current write output file
	marshaler codec.BinaryMarshaler // marshaler used for re-marshalling the ABCI messages to write them out to the destination files
	fileLock *sync.Mutex // mutex to sync access to the dst file since
	// NOTE: I suspect this lock is unnecessary since everything above the FileStreamingService occurs synchronously,
	// e.g. we dont need to worry about a kv pair being sent to the srcChan at the same time a new file is being generated
}
```

This streaming service uses a single instance of a simple intermediate `io.Writer` as the underlying `io.Writer` for the single `StoreKVPairWriteListener`,
collecting the KV pairs from every KVStore synchronously off of the same channel, and then writing
them out to the destination file generated for the current ABCI stage (as outlined above, with optional prefixes to avoid potential naming collisions
across separate instances).

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
func NewFileStreamingService(writeDir, filePrefix string, storeKeys []sdk.StoreKey, m codec.BinaryMarshaler) (*FileStreamingService, error) {
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
		listeners: listener
		srcChan: listenChan,
		filePrefix: filePrefix,
		writeDir: writeDir,
		marshaler: m,
		fileLock: new(sync.Mutex),
	}, nil
}

// Listeners returns the StreamingService's underlying WriteListeners, use for registering them with the BaseApp
func (fss *FileStreamingService) Listeners() map[sdk.StoreKey][]storeTypes.WriteListener {
	return fss.listeners
}

func (fss *FileStreamingService) BeginBlockReq(req abci.RequestBeginBlock) {
	// lock
	// close the file currently at fss.dstFile
	// update fss.dstFile with a new file generated using the begin block request info, per the naming schema
	// marshall the request and write it at the head of the file
	// unlock
	// now all kv pair writes go to the new file
}

func (fss *FileStreamingService) BeginBlockRes(res abci.ResponseBeginBlock) {
	// lock
	// marshall the response and write it to the tail of the current fss.dstFile
	// unlock
}

func (fss *FileStreamingService) EndBlockReq(req abci.RequestEndBlock) {
	// lock
	// close the file currently at fss.dstFile
	// update fss.dstFile with a new file generated using the end block request info, per the naming schema
	// marshall the request and write it at the head of the file
	// unlock
	// now all kv pair writes go to the new file
}

func (fss *FileStreamingService) EndBlockRes(res abci.ResponseEndBlock) {
	// lock
	// marshall the response and write it at the tail of the current fss.dstFile
	// unlock
}

func (fss *FileStreamingService) DeliverTxReq(req abci.RequestDeliverTx) {
	// lock
	// close the file currently at fss.dstFile
	// update fss.dstFile with a new file generated using the deliver tx request info, per the naming schema
	// marshall the request and then write it at the head of that file
	// unlock
	// now all writes go to the new file
}

func (fss *FileStreamingService) DeliverTxRes(res abci.ResponseDeliverTx) {
	// lock
	// marshall the response and write it at the tail of the current fss.dstFile
	// unlock
}

// Stream spins up a goroutine select loop which awaits length-prefixed binary encoded KV pairs to write out to the
// current destination file or a quit signal to shutdown the service
func (fss *FileStreamingService) Stream(wg *sync.WaitGroup, quitChan <-chan struct{}) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quitChan:
				return
                        case by := <-fss.srcChan:
                        	fss.fileLock.Wait()
                        	fss.dstFile.Write(by)
			}
		}
	}()
}
```

Writing to a file is the simplest approach for streaming the data out to consumers.
This approach also provide the advantages of being persistent and durable, and the files can be read directly,
or an auxiliary streaming services can read from the files and serve the data over a remote interface.

#### File pruning

Without pruning the number of files can grow indefinitely, this will need to be managed by
the developer in an application or even module-specific manner.
The file naming schema facilitates pruning by block number and/or ABCI message.

### Configuration

We will provide detailed documentation on how to configure a `FileStreamingService` from within an app's `AppCreator`,
using the provided `AppOptions` and TOML configuration fields.

#### BaseApp registration

We will add a new method to the `BaseApp` to enable the registration of `StreamingService`s:

```go
// RegisterStreamingService is used to register a streaming service with the BaseApp
func (app *BaseApp) RegisterStreamingService(s StreamingService) {
	// set the listeners for each StoreKey
	for key, lis := range s.Listeners() {
		app.cms.SetListeners(key, lis)
    }
    // register the streaming service within the BaseApp
    // BaseApp will pass BeginBlock, DeliverTx, and EndBlock requests and responses to the streaming services to update their ABCI context
    app.streamingServices = append(app.streamingServices, serv)
}
```

We will also modify the `BeginBlock`, `EndBlock`, and `DeliverTx` methods to pass ABCI requests and responses to any `StreamingServices` registered
with the `BaseApp`.


```go
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {
	defer telemetry.MeasureSince(time.Now(), "abci", "begin_block")

	if app.cms.TracingEnabled() {
		app.cms.SetTracingContext(sdk.TraceContext(
			map[string]interface{}{"blockHeight": req.Header.Height},
		))
	}

	if err := app.validateHeight(req); err != nil {
		panic(err)
	}
	
	// NEW CODE HERE
	// Update any registered streaming services with the new RequestBeginBlock message
	for _, streamingService := range app.streamingServices {
		streamingSerice.BeginBlockReq(req)
	}

	// Initialize the DeliverTx state. If this is the first block, it should
	// already be initialized in InitChain. Otherwise app.deliverState will be
	// nil, since it is reset on Commit.
	if app.deliverState == nil {
		app.setDeliverState(req.Header)
	} else {
		// In the first block, app.deliverState.ctx will already be initialized
		// by InitChain. Context is now updated with Header information.
		app.deliverState.ctx = app.deliverState.ctx.
			WithBlockHeader(req.Header).
			WithBlockHeight(req.Header.Height)
	}

	// add block gas meter
	var gasMeter sdk.GasMeter
	if maxGas := app.getMaximumBlockGas(app.deliverState.ctx); maxGas > 0 {
		gasMeter = sdk.NewGasMeter(maxGas)
	} else {
		gasMeter = sdk.NewInfiniteGasMeter()
	}

	app.deliverState.ctx = app.deliverState.ctx.WithBlockGasMeter(gasMeter)

	if app.beginBlocker != nil {
		res = app.beginBlocker(app.deliverState.ctx, req)
		res.Events = sdk.MarkEventsToIndex(res.Events, app.indexEvents)
	}
	// set the signed validators for addition to context in deliverTx
	app.voteInfos = req.LastCommitInfo.GetVotes() 
	
	// NEW CODE HERE
	// Update any registered streaming services with the new ResponseBeginBlock message
	for _ streamingService := range app.streamingServices {
		streamingService.BeginBlockRes(res)
	}
	
	return res
}
```

```go
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {
	defer telemetry.MeasureSince(time.Now(), "abci", "end_block")

	if app.deliverState.ms.TracingEnabled() {
		app.deliverState.ms = app.deliverState.ms.SetTracingContext(nil).(sdk.CacheMultiStore)
	}

	// NEW CODE HERE
	// Update any registered streaming services with the new RequestEndBlock message
	for _, streamingService := range app.streamingServices {
		streamingService.EndBlockReq(req)
	}

	if app.endBlocker != nil {
		res = app.endBlocker(app.deliverState.ctx, req)
		res.Events = sdk.MarkEventsToIndex(res.Events, app.indexEvents)
	}

	if cp := app.GetConsensusParams(app.deliverState.ctx); cp != nil {
		res.ConsensusParamUpdates = cp
	}
	
	// NEW CODE HERE
	// Update any registered streaming services with the new RequestEndBlock message 
	for _, streamingService := range app.streamingServices {
		streamingService.EndBlockRes(res)
	}
	
	return res
}
```

```go
func (app *BaseApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	defer telemetry.MeasureSince(time.Now(), "abci", "deliver_tx")

	gInfo := sdk.GasInfo{}
	resultStr := "successful"

	defer func() {
		telemetry.IncrCounter(1, "tx", "count")
		telemetry.IncrCounter(1, "tx", resultStr)
		telemetry.SetGauge(float32(gInfo.GasUsed), "tx", "gas", "used")
		telemetry.SetGauge(float32(gInfo.GasWanted), "tx", "gas", "wanted")
	}() 
	
	// NEW CODE HERE
	// Update any registered streaming services with the new RequestEndBlock message 
	for _, streamingService := range app.streamingServices {
		streamingService.DeliverTxReq(req)
	}

	gInfo, result, err := app.runTx(runTxModeDeliver, req.Tx)
	if err != nil {
		resultStr = "failed"
		return sdkerrors.ResponseDeliverTx(err, gInfo.GasWanted, gInfo.GasUsed, app.trace)
	}

	res := abci.ResponseDeliverTx{
		GasWanted: int64(gInfo.GasWanted), // TODO: Should type accept unsigned ints?
		GasUsed:   int64(gInfo.GasUsed),   // TODO: Should type accept unsigned ints?
		Log:       result.Log,
		Data:      result.Data,
		Events:    sdk.MarkEventsToIndex(result.Events, app.indexEvents),
	}
	
	// NEW CODE HERE
	// Update any registered streaming services with the new RequestEndBlock message 
	for _, streamingService := range app.streamingServices {
		streamingService.DeliverTxRes(res)
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
        keys = ["list", "of", "store", "keys", "we", "want", "to", "expose", "for", "this", "streamer"]
        writeDir = "path to the write directory"
        prefix = "optional prefix to prepend to the generated file names"
```

We will also provide a mapping of the TOML `store.streamers` "file" configuration option to a helper functions for constructing the specified
streaming service. In the future, as other streaming services are added, their constructors will be added here as well.

```go
// StreamingServiceConstructor is used to construct a streaming service
type StreamingServiceConstructor func(opts servertypes.AppOptions, keys []sdk.StoreKey) (StreamingService, error)

// StreamingServiceType enum for specifying the type of StreamingService
type StreamingServiceType int

const (
	Unknown StreamingServiceType = iota
	File
	// add more in the future
)

// NewStreamingServiceType returns the StreamingServiceType corresponding to the provided name
func NewStreamingServiceType(name string) StreamingServiceType {
	switch strings.ToLower(name) {
	case "file", "f":
		return File
	default:
		return Unknown
	}
}

// String returns the string name of a StreamingServiceType
func (sst StreamingServiceType) String() string {
	switch sst {
	case File:
		return "file"
	default:
		return ""
	}
}

// StreamingServiceConstructorLookupTable is a mapping of StreamingServiceTypes to StreamingServiceConstructors
var StreamingServiceConstructorLookupTable = map[StreamingServiceType]StreamingServiceConstructor{
	File: FileStreamingConstructor,
}

// NewStreamingServiceConstructor returns the StreamingServiceConstructor corresponding to the provided name
func NewStreamingServiceConstructor(name string) (StreamingServiceConstructor, error) {
	ssType := NewStreamingServiceType(name)
	if ssType == Unknown {
		return nil, fmt.Errorf("unrecognized streaming service name %s", name)
	}
	if constructor, ok := StreamingServiceConstructorLookupTable[ssType]; ok {
		return constructor, nil
	}
	return nil, fmt.Errorf("streaming service constructor of type %s not found", ssType.String())
}

// FileStreamingConstructor is the StreamingServiceConstructor function for creating a FileStreamingService
func FileStreamingConstructor(opts servertypes.AppOptions, keys []sdk.StoreKey) (StreamingService, error) {
	filePrefix := cast.ToString(opts.Get("streamers.file.prefix"))
	fileDir := cast.ToString(opts.Get("streamers.file.writeDir"))
	return streaming.NewFileStreamingService(fileDir, filePrefix, keys), nil
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
		// get the store keys allowed to be exposed for this streaming service/state listeners
		exposeKeyStrs := cast.ToStringSlice(appOpts.Get(fmt.Sprintf("streamers.%s.keys", listenerName))
		exposeStoreKeys = make([]storeTypes.StoreKey, 0, len(exposeKeyStrs))
		for _, keyStr := range exposeKeyStrs {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
		// get the constructor for this listener name
		constructor, err := baseapp.NewStreamingServiceConstructor(listenerName)
		if err != nil {
			tmos.Exit(err.Error()) // or continue?
		}
		// generate the streaming service using the constructor, appOptions, and the StoreKeys we want to expose
		streamingService, err := constructor(appOpts, exposeStoreKeys)
		if err != nil {
			tmos.Exit(err.Error())
		}
		// register the streaming service with the BaseApp
		bApp.RegisterStreamingService(streamingService)
		// waitgroup and quit channel for optional shutdown coordination of the streaming service
		wg := new(sync.WaitGroup)
		quitChan := new(chan struct{}))
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
