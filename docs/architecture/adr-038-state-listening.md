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
// WriteListener interface for writing data out from a listenkv.Store
type WriteListener interface {
  // if value is nil then it was deleted
  OnWrite(key []byte, value []byte)
}
```

### Listener type
We will create a concrete implementation of the `WriteListener` interface in `store/types/listening.go` that gob
encodes and writes the key value pair into an underlying `io.Writer`.

```go
// GobWriteListener is used to configure listening to a KVStore using a gob encoding wrapper around an io.Writer
type GobWriteListener struct {
	writer              *gob.Encoder
}

// NewGobWriteListener wraps a WriteListenerWriter around an io.Writer
func NewGobWriteListener(w io.Writer) *GobWriteListener {
	return &GobWriteListener{
		writer: gob.NewEncoder(w),
	}
}

// KVPair for gob encoding
type KVPair struct {
	Key []byte
	Value []byte
}

// OnWrite satisfies WriteListener interface by writing out key-value gobs to the underlying io.Writer
func (l *Listener) OnWrite(key []byte, value []byte) {
	l.writer.Encode(KVPair{Key: key, Value: value})
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
func onWrite(listeners []types.WriteListener, key, value []byte) {
	for _, l := range listeners {
		l.OnWrite(key, value)
	}
}
```

### MultiStore interface updates
We will update the `MultiStore` interface to allow us to wrap a set of listeners around a specific `KVStore`.
Additionally, we will update the `CacheWrap` and `CacheWrapper` interfaces to enable listening in the caching layer, and add a `MultiStore` method
to turn on or off this cache listening.

```go
type MultiStore interface {
	...

	// ListeningEnabled returns if listening is enabled for the KVStore belonging the provided StoreKey
	ListeningEnabled(key StoreKey) bool

	// SetListeners sets the WriteListeners for the KVStore belonging to the provided StoreKey
	SetListeners(key StoreKey, listeners []WriteListener)

	// CacheListening enables or disables KVStore listening at the cache layer
	CacheListening(listen bool)
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
to wrap the returned `KVStore` with a `listenkv.Store` if listening is turned on.

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

We will also adjust the `cachemulti` constructor methods and the `rootmulti` `CacheMultiStore` method to enable listening
in the cache layer when `CacheListening` is turned on.

```go
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}
	var cacheListeners map[types.StoreKey][]types.WriteListener
	if rs.cacheListening {
		cacheListeners = rs.listeners
	}
	return cachemulti.NewStore(rs.db, stores, rs.keysByName, rs.traceWriter, rs.traceContext, cacheListeners)
}
```

### Exposing the data 
We will introduce and document mechanisms for exposing data from the above listeners to external consumers.

#### Writing to file
We will document and provide examples of how to configure a listener to write out to a file.
No new type implementation will be needed, an `os.File` can be used as the underlying `io.Writer` for a `GobWriteListener`.

Writing to a file is the simplest approach for streaming the data out to consumers.
This approach also provide the advantages of being persistent and durable, and the files can be read directly
or an auxiliary streaming services can tail the files and serve the data remotely.

Without pruning the file size can grow indefinitely, this will need to be managed by
the developer in an application or even module-specific manner.

#### Writing to gRPC stream
We will implement and document an `io.Writer` type for exposing our listeners over a gRPC server stream.

Writing to a gRPC stream gRPC will allow us to expose the data over the standard gRPC interface.
This interface can be exposed directly to consumers or we can implement a message queue or secondary streaming service on top.
Using gRPC will provide us with all of the regular advantages of gRPC and protobuf: versioning guarantees, client side code generation, and interoperability with the many gRPC plugins and auxillary services.

Proceeding through a gRPC intermediate will provide additional overhead, in most cases this is not expected to be rate limiting but in
instances where it is the developer can implement a more performant streaming mechanism for state listening.

### Configuration
We will provide detailed documentation on how to configure the state listeners and their external streaming services from within an app's `AppCreator`,
using the provided `AppOptions`. We will add two methods to the `BaseApp` to enable this configuration:

```go
// SetCommitMultiStoreListeners sets the KVStore listeners for the provided StoreKey
func (app *BaseApp) SetCommitMultiStoreListeners(key sdk.StoreKey, listeners []storeTypes.WriteListener) {
	app.cms.SetListeners(key, listeners)
}

// SetCacheListening turns on or off listening at the cache layer
func (app *BaseApp) SetCacheListening(listening bool) {
	app.cms.CacheListening(listening)
}
```

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

	// configure state listening capabilities using AppOptions
	if cast.ToBool(appOpts.Get("simApp.listening")) {
		writeDir := filepath.Clean(cast.ToString(appOpts.Get("simApp.listening.writeDir")))
		for _, key := range keys {
			loadListener(bApp, writeDir, key)
		}
		for _, key := range tkeys {
			loadListener(bApp, writeDir, key)
		}
		for _, key := range memKeys {
			loadListener(bApp, writeDir, key)
		}
		bApp.SetCacheListening(cast.ToBool(appOpts.Get("simApp.cacheListening")))
	}
	
	...

	return app
}

// loadListener creates and adds to the BaseApp a listener that writes out to a file in the provided write directory
// The file is named after the StoreKey for the KVStore it listens to 
func loadListener(bApp *baseapp.BaseApp, writeDir string, key sdk.StoreKey) {
	writePath := filepath.Join(writeDir, key.Name())
	// TODO: how to handle graceful file closure?
	fileHandler, err := os.OpenFile(writePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		tmos.Exit(err.Error())
	}
	// using single io.Writer based listener
	listener := storeTypes.NewGobWriteListener(fileHandler)
	bApp.SetCommitMultiStoreListeners(key, []storeTypes.Listening{listener})
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
