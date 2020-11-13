# ADR 038: KVStore state listening

## Changelog

- 11/23/2020: Initial draft

## Status

Proposed

## Abstract

This ADR defines a set of changes to enable state change listening of individual KVStores.

## Context

Currently, KVStore data can be remotely accessed through [Queries](https://github.com/cosmos/cosmos-sdk/blob/master/docs/building-modules/messages-and-queries.md#queries) which proceed through Tendermint and the ABCI.
In addition to these request/response queries, it would be beneficial to have a means of listening to state changes as they occur in real time.

## Decision

We will modify the MultiStore interface and its concrete (`basemulti` and `cachemulti`) implementations and introduce a new `listenkv.Store` to allow listening to specific state changes in underlying KVStores and routing the output to consumers.
We will also introduce two approaches for exposing the data to consumers: writing to files and writing to a gRPC stream.

### Listening interface
In a new file- `store/types/listening.go`- we will create a `Listening` interface for streaming out an allowed subset of state changes from a KVStore.
The interface can be backed by a simple wrapper around any underlying `io.Writer`.

```go
// Listening interface comprises the methods needed to filter and stream data out from a KVStore
type Listening interface {
	io.Writer // the underlying io.Writer
	Allowed(op Operation, key []byte) bool // method used to check if the Listener is allowed to listen to a specific state change
	TraceContext() TraceContext // method to access this Listener's TraceContext
}
```

We will lift some of the private types currently being used by the `tracekv.Store` into public types at a new location- `store/types/tracing.go`- for use in the `Listening` interface
```go
type Operation string

const (
	WriteOp     Operation = "write"
	ReadOp      Operation = "read"
	DeleteOp    Operation = "delete"
	IterKeyOp   Operation = "iterKey"
	IterValueOp Operation = "iterValue"
)

// TraceOperation implements a traced KVStore operation
type TraceOperation struct {
	Operation Operation              `json:"operation"`
	Key       string                 `json:"key"`
	Value     string                 `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
}
```

### Listener type
We will create a concrete implementation of the `Listening` interface in `store/types/listening.go`.
This implementation will be configurable with a list of allowed `Operation`s, a set of whitelisted keys and prefixes, and a set of blacklisted keys and prefixes.

```go

// Listener is used to configure listening on specific keys of a KVStore
type Listener struct {
	writer              io.Writer
	context             TraceContext
	allowedOperations   map[Operation]struct{} // The operations which this listener is allowed to listen to
	whitelistedKeys     [][]byte               // Keys explicitly allowed to be listened to
	blacklistedKeys     [][]byte               // Keys explicitly disallowed to be listened to
	whitelistedPrefixes [][]byte               // Key prefixes explicitly allowed to be listened to
	blacklistedPrefixes [][]byte               // Key prefixes explicitly disallowed to be listened to
}

...

// GetContext satisfies Listening interface
func (l *Listener) GetContext() TraceContext {
	return l.context
}

// Write satisfies Listening interface
// it wraps around the underlying writer interface
func (l *Listener) Write(b []byte) (int, error) {
	return l.writer.Write(b)
}

// Allowed satisfies Listening interface
// it returns whether or not the Listener is allowed to listen to the provided operation at the provided key
func (l *Listener) Allowed(op Operation, key []byte) bool {
	// first check if the operation is allowed
	if _, ok := l.allowedOperations[op]; !ok {
		return false
	}
	// if there are no keys or prefixes in the whitelists then every key is allowed (unless disallowed in blacklists)
	// if there are whitelisted keys or prefixes then only the keys which conform are allowed (unless disallowed in blacklists)
	allowed := true
	if len(l.whitelistedKeys) > 0 || len(l.whitelistedPrefixes) > 0 {
		allowed = listsContain(l.whitelistedKeys, l.whitelistedPrefixes, key)
	}
	return allowed && !listsContain(l.blacklistedKeys, l.blacklistedPrefixes, key)
}

func listsContain(keys, prefixes [][]byte, key []byte) bool {
	for _, k := range keys {
		if bytes.Equal(key, k) {
			return true
		}
	}
	for _, p := range prefixes {
		if bytes.HasPrefix(key, p) {
			return true
		}
	}
	return false
}
```

### ListenKVStore
We will create a new `Store` type `listenkv.Store` that the `MultiStore` wraps around a `KVStore` to enable state listening.
This is closely modeled after the `tracekv.Store` with the primary difference being that we can configure the `Store` with a set of `Listening` types
which direct the streaming of only certain allowed subsets of keys and/or operations to specific `io.Writer` destinations.

```go
// Store implements the KVStore interface with listening enabled.
// Operations are traced on each core KVStore call and written to any of the
// underlying listeners with the proper key and operation permissions
type Store struct {
    parent    types.KVStore
    listeners []types.Listener
}

// NewStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewStore(parent types.KVStore, listeners []types.Listener) *Store {
	return &Store{parent: parent, listeners: listeners}
}

...

// Get implements the KVStore interface. It traces a read operation and
// delegates a Get call to the parent KVStore.
func (tkv *Store) Get(key []byte) []byte {
	value := tkv.parent.Get(key)

	writeOperation(tkv.listeners, types.ReadOp, key, value)
	return value
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (tkv *Store) Set(key []byte, value []byte) {
	types.AssertValidKey(key)
	writeOperation(tkv.listeners, types.WriteOp, key, value)
	tkv.parent.Set(key, value)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (tkv *Store) Delete(key []byte) {
	writeOperation(tkv.listeners, types.DeleteOp, key, nil)
	tkv.parent.Delete(key)
}

// Has implements the KVStore interface. It delegates the Has call to the
// parent KVStore.
func (tkv *Store) Has(key []byte) bool {
	return tkv.parent.Has(key)
}

...

// writeOperation writes a KVStore operation to the underlying io.Writer of
// every listener that has permissions to listen to that operation at the given key
// The TraceOperation is JSON-encoded with the `key` and `value` fields as base64 encoded strings
func writeOperation(listeners []types.Listener, op types.Operation, key, value []byte) {
	traceOp := types.TraceOperation{
		Operation: op,
		Key:       base64.StdEncoding.EncodeToString(key),
		Value:     base64.StdEncoding.EncodeToString(value),
	}
	for _, l := range listeners {
		if !l.Allowed(op, key) {
			continue
		}
		traceOp.Metadata = l.Context
		raw, err := json.Marshal(traceOp)
		if err != nil {
			panic(errors.Wrap(err, "failed to serialize listen operation"))
		}
		if _, err := l.Writer.Write(raw); err != nil {
			panic(errors.Wrap(err, "failed to write listen operation"))
		}
		io.WriteString(l.Writer, "\n")
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

	// SetListeners sets the listener set for the KVStore belonging to the provided StoreKey
	SetListeners(key StoreKey, listeners []Listener)

	// CacheListening enables or disables KVStore listening at the cache layer
	CacheListening(listen bool)
}
```

```go
type CacheWrap interface {
    ...

	// CacheWrapWithListeners recursively wraps again with listening enabled
	CacheWrapWithListeners(listeners []Listener) CacheWrap
}

type CacheWrapper interface {
	...

	// CacheWrapWithListeners recursively wraps again with listening enabled
	CacheWrapWithListeners(listeners []Listener) CacheWrap
}
```

### MultiStore implementation updates
We will modify all of the Stores and MultiStores to satisfy these new interfaces, and adjust the `rootmulti` MultiStore's `GetKVStore` method
to enable wrapping the returned `KVStore` with the `listenkv.Store`.

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

We will also adjust the `cachemulti` constructor methods and the `rootmulti` `CacheMultiStore` method to enable cache listening when `CacheListening` is turned on.

```go
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}
	var cacheListeners map[types.StoreKey][]types.Listener
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
No new type implementation is needed, a `os.File` can be used as the underlying `io.Writer` for a listener.
Writing to a file is the simplest approach for streaming the data out to consumers.
This approach also provide the advantages of being persistent and durable.
The files can be read directly or an auxiliary streaming services can tail the files and serve the data remotely.
Without pruning the file size can grow indefinitely, this will need to be managed by
the developer in an application or even module-specific manner.

#### Writing to gRPC stream
We will implement a `io.Writer` type for exposing our listeners over a gRPC server stream.
Writing to a gRPC stream gRPC allows us to expose the data over the standard gRPC interface.
This interface can be exposed directly to consumers or we can implement a message queue or streaming service logic on top.
Using gRPC provides us with all of the regular advantages of gRPC and protobuf: versioning guarantees, client side code generation, and interoperability with the many gRPC plugins and auxillary services.
Proceeding through a gRPC intermediate will provide additional overhead, in most cases this is not expected to be rate limiting but in
instances where it is the developer can implement a more performant streaming mechanism for state listening.

### Configuration
We will provide detailed documentation for how to configure the state listeners and their external streaming services from within an app's `AppCreator`,
using the provided `AppOptions`.

e.g. SimApp with simple state streaming to files:

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
	// using single listener with all operations and keys permitted
	listener := storeTypes.NewDefaultStateListener(fileHandler, nil)
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
