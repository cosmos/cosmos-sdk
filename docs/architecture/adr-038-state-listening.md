# ADR 038: KVStore state listening

## Changelog

- 11/23/2020: Initial draft

## Status

Proposed

## Abstract

This ADR defines a set of changes to enable listening to state changes of individual KVStores and exposing these data to consumers.

## Context

Currently, KVStore data can be remotely accessed through [Queries](https://github.com/cosmos/cosmos-sdk/blob/master/docs/building-modules/messages-and-queries.md#queries) which proceed through Tendermint and the ABCI.
In addition to these request/response queries, it would be beneficial to have a means of listening to state changes as they occur in real time.

## Decision

We will modify the `MultiStore` interface and its concrete (`rootmulti` and `cachemulti`) implementations and introduce a new `listenkv.Store` to allow listening to state changes in underlying KVStores.
We will also introduce the tooling for writing these state changes out to a file.

### Listening interface

In a new file- `store/types/listening.go`- we will create a `WriteListener` interface for streaming out state changes from a KVStore.

```go
// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
  // if value is nil then it was deleted
  OnWrite(key []byte, value []byte)
}
```

### Listener type

We will create two concrete implementation of the `WriteListener` interface in `store/types/listening.go`.

One that writes out length-prefixed key-value pairs to an underlying `io.Writer`:

```go
// PrefixWriteListener is used to configure listening to a KVStore by writing out big endian length-prefixed
// key-value pairs to an io.Writer
type PrefixWriteListener struct {
	writer io.Writer
	prefixBuf [6]byte
}

// NewPrefixWriteListener wraps a PrefixWriteListener around an io.Writer
func NewPrefixWriteListener(w io.Writer) *PrefixWriteListener {
	return &PrefixWriteListener{
		writer: w,
	}
}

// OnWrite satisfies the WriteListener interface by writing out big endian length-prefixed key-value pairs
// to an underlying io.Writer
// The first two bytes of the prefix encode the length of the key
// The last four bytes of the prefix encode the length of the value
// This WriteListener makes two assumptions
// 1) The key is no longer than 1<<16 - 1
// 2) The value is no longer than 1<<32 - 1
func (swl *PrefixWriteListener) OnWrite(key []byte, value []byte) {
	keyLen := len(key)
	valLen := len(key)
	if keyLen > math.MaxUint16 || valLen > math.MaxUint32 {
		return
	}
	binary.BigEndian.PutUint16(l.prefixBuf[:2], uint16(keyLen))
	binary.BigEndian.PutUint32(l.prefixBuf[2:], uint32(valLen))
	l.writer.Write(l.prefixBuf[:])
	l.writer.Write(key)
	l.writer.Write(value)
}
```

And one that writes out newline-delineated key-length-prefixed key-value pairs to an underlying io.Writer:

```go
// NewlineWriteListener is used to configure listening to a KVStore by writing out big endian key-length-prefixed and
// newline delineated key-value pairs to an io.Writer
type NewlineWriteListener struct {
	writer              io.Writer
	keyLenBuf           [2]byte
}

// NewNewlineWriteListener wraps a StockWriteListener around an io.Writer
func NewNewlineWriteListener(w io.Writer) *NewlineWriteListener {
	return &NewlineWriteListener{
		writer: w,
	}
}

var newline = []byte("\n")

// OnWrite satisfies WriteListener interface by writing out newline delineated big endian key-length-prefixed key-value
// pairs to the underlying io.Writer
// The first two bytes encode the length of the key
// Separate key-value pairs are newline delineated
// This WriteListener makes three assumptions
// 1) The key is no longer than 1<<16 - 1
// 2) The value and keys contain no newline characters
func (l *NewlineWriteListener) OnWrite(key []byte, value []byte) {
	keyLen := len(key)
	if keyLen > math.MaxUint16 {
		return
	}
	binary.BigEndian.PutUint16(l.keyLenBuf[:], uint16(keyLen))
	l.writer.Write(e.keyLenBuf[:])
	l.writer.Write(key)
	l.writer.Write(value)
	l.writer.Write(newline)
}
```

The former makes no assumptions about the presence of newline characters in keys or values, but values
must be no longer than 1<<32 - 1. The latter assumes newlines are not present in keys or values but can support any length
of value. Both assume keys are no longer than 1<<16 - 1. Newline delineation improves readability by enabling a consumer to orient
themselves at the start of a key-value pair at any point in the stream (e.g. tail a file), without character delineation a consumer must start
at the beginning of the stream and not lose track of their position in the stream.

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
}

// NewStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewStore(parent types.KVStore, listeners []types.WriteListener) *Store {
	return &Store{parent: parent, listeners: listeners}
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (tkv *Store) Set(key []byte, value []byte) {
	types.AssertValidKey(key)
	onWrite(tkv.listeners, key, value)
	tkv.parent.Set(key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (tkv *Store) Delete(key []byte) {
	onWrite(tkv.listeners, key, nil)
	tkv.parent.Delete(key)
}

// onWrite writes a KVStore operation to all of the WriteListeners
// Note: write out in a goroutine
func onWrite(listeners []types.WriteListener, key, value []byte) {
	for _, l := range listeners {
		l.OnWrite(key, value)
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
	CacheWrapWithListeners(listeners []WriteListener) CacheWrap
}

type CacheWrapper interface {
	...

	// CacheWrapWithListeners recursively wraps again with listening enabled
	CacheWrapWithListeners(listeners []WriteListener) CacheWrap
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
		store = listenkv.NewStore(store, rs.listeners[key])
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
	Listeners() map[sdk.StoreKey][]storeTypes.WriteListener // returns the streaming service's listeners for the BaseApp to register
	BeginBlockReq(req abci.RequestBeginBlock) // update the streaming service with the latest RequestBeginBlock message
	BeginBlockResres abci.ResponseBeginBlock) // update the steaming service with the latest ResponseBeginBlock message
	EndBlockReq(req abci.RequestEndBlock) // update the steaming service with the latest RequestEndBlock message
	EndBlockRes(res abci.ResponseEndBlock) // update the steaming service with the latest ResponseEndBlock message
	DeliverTxReq(req abci.RequestDeliverTx) // update the steaming service with the latest RequestDeliverTx message
    DeliverTxRes(res abci.ResponseDeliverTx) // update the steaming service with the latest ResponseDeliverTx message
}
```

#### Writing state changes to files

We will introduce an implementation of `StreamingService` which writes state changes out to a file.

```go
// FileStreamingService is a concrete implementation of StreamingService that writes state changes out to a file
type FileStreamingService struct {
	listeners map[sdk.StoreKey][]storeTypes.WriteListener // the listeners that will be initialized with BaseApp
	writeDir string
	filePrefix string
	fileSuffix string
	dst *os.File // the current write output file
}

```

Writing to a file is the simplest approach for streaming the data out to consumers.
This approach also provide the advantages of being persistent and durable, and the files can be read directly,
or an auxiliary streaming services can tail the files and serve the data remotely.

#### File pruning

Without pruning the file size can grow indefinitely, this will need to be managed by
the developer in an application or even module-specific manner (e.g. log rotation).

### Configuration

We will provide detailed documentation on how to configure the state listeners and the file streaming service from within an app's `AppCreator`,
using the provided `AppOptions`.

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

We will also modify the `BeginBlock`, `EndBlock`, and `DeliverTx` methods to pass messages and responses to any `StreamingServices` registered
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
	
	// Update any registered streaming services with the new RequestEndBlock message 
	for _, streamingService := range app.streamingServices {
		streamingService.DeliverTxRes(res)
	}
	
	return res
}
```

#### TOML Configuration

We will provide a standard TOML configuration options for configuring a `FileStreamingService` for specific `Store`s.
Note: the actual namespace is TBD.

```toml
[store]
    streamers = [ # if len(streamers) > 0 we are streamers
        "file",
    ]

[streamers]
    [streamers.file]
        keys = ["list", "of", "store", "keys", "we", "want", "to", "expose", "for", "this", "streamer"]
        writeDir = "path to the write directory"
        filePrefix = "optional string to prefix the file names with"
        fileSuffix = "optional string to suffix the file names with"
```

We will also provide a mapping of the TOML `store.streamer` configuration options to helper functions for constructing the specified
streaming service.

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
	...
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
		// get the store keys allowed to be exposed for this listener
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

- Introduces additional- but optional- complexity to configuring and running a cosmos app
- If an application developer opts to use these features to expose data, they need to be aware of the ramifications/risks of that data exposure as it pertains to the specifics of their application
