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
We will also introduce two approaches for exposing the data to external consumers: writing to files and writing to a gRPC stream.

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
of value. Both assume keys are no longer than 1<<16 - 1. Newline delineation improves durability by enabling a consumer to orient
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

We will introduce and document mechanisms for exposing data from the above listeners to external consumers.

#### Writing to file

We will document and provide examples of how to configure a listener to write out to a file.
No new type implementation will be needed, an `os.File` can be used as the underlying `io.Writer` for a `GobWriteListener`.

Writing to a file is the simplest approach for streaming the data out to consumers.
This approach also provide the advantages of being persistent and durable, and the files can be read directly,
or an auxiliary streaming services can tail the files and serve the data remotely.

Without pruning the file size can grow indefinitely, this will need to be managed by
the developer in an application or even module-specific manner (e.g. log rotation).

#### Writing to gRPC stream

We will implement and document an `io.Writer` type for exposing our listeners over a gRPC server stream.

Writing to a gRPC stream gRPC will allow us to expose the data over the standard gRPC interface.
This interface can be exposed directly to consumers, or we can implement a message queue or secondary streaming service on top.
Using gRPC will provide us with all the regular advantages of gRPC and protobuf: versioning guarantees,
client side code generation, and interoperability with the many gRPC plugins and auxiliary services.

Proceeding through a gRPC intermediate will provide additional overhead, in most cases this is not expected to be rate limiting but in
instances where it is the developer can implement a more performant streaming mechanism for state listening.

### Configuration

We will provide detailed documentation on how to configure the state listeners and their external streaming services from within an app's `AppCreator`,
using the provided `AppOptions`. We will add a new method to the `BaseApp` to enable this configuration:

```go
// SetCommitMultiStoreListeners sets the KVStore listeners for the provided StoreKey
func (app *BaseApp) SetCommitMultiStoreListeners(key sdk.StoreKey, listeners []storeTypes.WriteListener) {
	app.cms.SetListeners(key, listeners)
}
```

### TOML Configuration

We will provide standard TOML configuration options for configuring the state listeners and their external streaming services.
Note: the actual namespace is TBD.

```toml
[store]
    listeners = [ # if len(listeners) > 0 we are listening
        "file",
        "grpc"
    ]
```

We will also provide a mapping of these TOML configuration options to helper functions for setting up the specified
streaming service.

```go
// StreamingServiceConstructor is used to construct and load a WriteListener onto the provided BaseApp and expose it over a streaming service
type StreamingServiceConstructor func(bApp *BaseApp, opts servertypes.AppOptions, keys []sdk.StoreKey) error

// StreamingServiceType enum for specifying the type of StreamingService
type StreamingServiceType int

const (
	Unknown StreamingServiceType = iota
	File
	GRPC
)

// NewStreamingServiceType returns the StreamingServiceType corresponding to the provided name
func NewStreamingServiceType(name string) StreamingServiceType {
	switch strings.ToLower(name) {
	case "file", "f":
		return File
	case "grpc":
		return GRPC
	default:
		return Unknown
	}
}

// String returns the string name of a StreamingServiceType
func (sst StreamingServiceType) String() string {
	switch sst {
	case File:
		return "file"
	case GRPC:
		return "grpc"
	default:
		return ""
	}
}

// StreamingServiceConstructorLookupTable is a mapping of StreamingServiceTypes to StreamingServiceConstructors
var StreamingServiceConstructorLookupTable = map[StreamingServiceType]StreamingServiceConstructor{
	File: FileStreamingConstructor,
	GRPC: GRPCStreamingConstructor,
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

// FileStreamingConstructor is the StreamingServiceConstructor function for writing out to a file
func FileStreamingConstructor(bApp *BaseApp, opts servertypes.AppOptions, keys []sdk.StoreKey) error {
	...
}

// GRPCStreamingConstructor is the StreamingServiceConstructor function for writing out to a gRPC stream
func GRPCStreamingConstructor(bApp *BaseApp, opts servertypes.AppOptions, keys []sdk.StoreKey) error {
	...
}
```

### Example configuration

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
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	
	// collect the keys for the stores we wish to expose 
	storeKeys := make([]storeTypes.StoreKey, 0, len(keys))
	for _, key := range keys {
		storeKeys = append(storeKeys, key)
	}
	// configure state listening capabilities using AppOptions
	listeners := cast.ToStringSlice(appOpts.Get("store.listeners"))
	for _, listenerName := range listeners {
		constructor, err := baseapp.NewStreamingServiceConstructor(listenerName)
		if err != nil {
			tmos.Exit(err.Error()) // or continue?
		}
		if err := constructor(bApp, appOpts, storeKeys); err != nil {
			tmos.Exit(err.Error())
		}
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
