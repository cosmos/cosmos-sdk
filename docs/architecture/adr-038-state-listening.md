# ADR 038: KVStore state listening

## Changelog

* 11/23/2020: Initial draft
* 10/06/2022: Introduce plugin system based on hashicorp/go-plugin
* 10/14/2022:
    * Add `ListenCommit`, flatten the state writes in a block to a single batch.
    * Remove listeners from cache stores, should only listen to `rootmulti.Store`.
    * Remove `HaltAppOnDeliveryError()`, the errors are propagated by default, the implementations should return nil if don't want to propagate errors.
* 26/05/2023: Update with ABCI 2.0

## Status

Proposed

## Abstract

This ADR defines a set of changes to enable listening to state changes of individual KVStores and exposing these data to consumers.

## Context

Currently, KVStore data can be remotely accessed through [Queries](https://github.com/cosmos/cosmos-sdk/blob/main/docs/build/building-modules/02-messages-and-queries.md#queries)
which proceed either through Tendermint and the ABCI, or through the gRPC server.
In addition to these request/response queries, it would be beneficial to have a means of listening to state changes as they occur in real time.

## Decision

We will modify the `CommitMultiStore` interface and its concrete (`rootmulti`) implementations and introduce a new `listenkv.Store` to allow listening to state changes in underlying KVStores. We don't need to listen to cache stores, because we can't be sure that the writes will be committed eventually, and the writes are duplicated in `rootmulti.Store` eventually, so we should only listen to `rootmulti.Store`.
We will introduce a plugin system for configuring and running streaming services that write these state changes and their surrounding ABCI message context to different destinations.

### Listening

In a new file, `store/types/listening.go`, we will create a `MemoryListener` struct for streaming out protobuf encoded KV pairs state changes from a KVStore.
The `MemoryListener` will be used internally by the concrete `rootmulti` implementation to collect state changes from KVStores.

```go
// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the the same WriteListener across separate KVStores
	// delete bool indicates if it was a delete; true: delete, false: set
    OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool) error
}

// NewMemoryListener creates a listener that accumulate the state writes in memory.
func NewMemoryListener() *MemoryListener {
	return &MemoryListener{}
}

// OnWrite writes state change events to the internal cache
func (fl *MemoryListener) OnWrite(storeKey StoreKey, key []byte, value []byte, delete bool) {
	fl.stateCache = append(fl.stateCache, StoreKVPair{
		StoreKey: storeKey.Name(),
		Delete:   delete,
		Key:      key,
		Value:    value,
	})
}

// PopStateCache returns the current state caches and set to nil
func (fl *MemoryListener) PopStateCache() []StoreKVPair {
	res := fl.stateCache
	fl.stateCache = nil
	return res
}
```

We will also define a protobuf type for the KV pairs. In addition to the key and value fields this message
will include the StoreKey for the originating KVStore so that we can collect information from separate KVStores and determine the source of each KV pair.

```protobuf
message StoreKVPair {
  optional string store_key = 1; // the store key for the KVStore this pair originates from
  required bool set = 2; // true indicates a set operation, false indicates a delete operation
  required bytes key = 3;
  required bytes value = 4;
}
```

### ListenKVStore

We will create a new `Store` type `listenkv.Store` that the `rootmulti` store will use to wrap a `KVStore` to enable state listening.
We will configure the `Store` with a `MemoryListener` which will collect state changes for output to specific destinations.

```go
// Store implements the KVStore interface with listening enabled.
// Operations are traced on each core KVStore call and written to any of the
// underlying listeners with the proper key and operation permissions
type Store struct {
    parent    types.KVStore
    listener  *types.MemoryListener
    parentStoreKey types.StoreKey
}

// NewStore returns a reference to a new traceKVStore given a parent
// KVStore implementation and a buffered writer.
func NewStore(parent types.KVStore, psk types.StoreKey, listener *types.MemoryListener) *Store {
    return &Store{parent: parent, listener: listener, parentStoreKey: psk}
}

// Set implements the KVStore interface. It traces a write operation and
// delegates the Set call to the parent KVStore.
func (s *Store) Set(key []byte, value []byte) {
    types.AssertValidKey(key)
    s.parent.Set(key, value)
    s.listener.OnWrite(s.parentStoreKey, key, value, false)
}

// Delete implements the KVStore interface. It traces a write operation and
// delegates the Delete call to the parent KVStore.
func (s *Store) Delete(key []byte) {
    s.parent.Delete(key)
    s.listener.OnWrite(s.parentStoreKey, key, nil, true)
}
```

### MultiStore interface updates

We will update the `CommitMultiStore` interface to allow us to wrap a `Memorylistener` to a specific `KVStore`.
Note that the `MemoryListener` will be attached internally by the concrete `rootmulti` implementation.

```go
type CommitMultiStore interface {
    ...

    // AddListeners adds a listener for the KVStore belonging to the provided StoreKey
    AddListeners(keys []StoreKey)

    // PopStateCache returns the accumulated state change messages from MemoryListener
    PopStateCache() []StoreKVPair
}
```


### MultiStore implementation updates

We will adjust the `rootmulti` `GetKVStore` method to wrap the returned `KVStore` with a `listenkv.Store` if listening is turned on for that `Store`.

```go
func (rs *Store) GetKVStore(key types.StoreKey) types.KVStore {
    store := rs.stores[key].(types.KVStore)

    if rs.TracingEnabled() {
        store = tracekv.NewStore(store, rs.traceWriter, rs.traceContext)
    }
    if rs.ListeningEnabled(key) {
        store = listenkv.NewStore(store, key, rs.listeners[key])
    }

    return store
}
```

We will implement `AddListeners` to manage KVStore listeners internally and implement `PopStateCache`
for a means of retrieving the current state.

```go
// AddListeners adds state change listener for a specific KVStore
func (rs *Store) AddListeners(keys []types.StoreKey) {
	listener := types.NewMemoryListener()
	for i := range keys {
		rs.listeners[keys[i]] = listener
	}
}
```

```go
func (rs *Store) PopStateCache() []types.StoreKVPair {
	var cache []types.StoreKVPair
	for _, ls := range rs.listeners {
		cache = append(cache, ls.PopStateCache()...)
	}
	sort.SliceStable(cache, func(i, j int) bool {
		return cache[i].StoreKey < cache[j].StoreKey
	})
	return cache
}
```

We will also adjust the `rootmulti` `CacheMultiStore` and `CacheMultiStoreWithVersion` methods to enable listening in
the cache layer.

```go
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
    stores := make(map[types.StoreKey]types.CacheWrapper)
    for k, v := range rs.stores {
        store := v.(types.KVStore)
        // Wire the listenkv.Store to allow listeners to observe the writes from the cache store,
        // set same listeners on cache store will observe duplicated writes.
        if rs.ListeningEnabled(k) {
            store = listenkv.NewStore(store, k, rs.listeners[k])
        }
        stores[k] = store
    }
    return cachemulti.NewStore(rs.db, stores, rs.keysByName, rs.traceWriter, rs.getTracingContext())
}
```

```go
func (rs *Store) CacheMultiStoreWithVersion(version int64) (types.CacheMultiStore, error) {
 // ...

        // Wire the listenkv.Store to allow listeners to observe the writes from the cache store,
        // set same listeners on cache store will observe duplicated writes.
        if rs.ListeningEnabled(key) {
            cacheStore = listenkv.NewStore(cacheStore, key, rs.listeners[key])
        }

        cachedStores[key] = cacheStore
    }

    return cachemulti.NewStore(rs.db, cachedStores, rs.keysByName, rs.traceWriter, rs.getTracingContext()), nil
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

#### BaseApp Registration

We will add a new method to the `BaseApp` to enable the registration of `StreamingService`s:

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
type BaseApp struct {

    ...

    // abciListenersAsync for determining if abciListeners will run asynchronously.
    // When abciListenersAsync=false and stopNodeOnABCIListenerErr=false listeners will run synchronized but will not stop the node.
    // When abciListenersAsync=true stopNodeOnABCIListenerErr will be ignored.
    abciListenersAsync bool

    // stopNodeOnABCIListenerErr halts the node when ABCI streaming service listening results in an error.
    // stopNodeOnABCIListenerErr=true must be paired with abciListenersAsync=false.
    stopNodeOnABCIListenerErr bool
}
```

#### ABCI Event Hooks

We will modify the `FinalizeBlock` and `Commit` methods to pass ABCI requests and responses
to any streaming service hooks registered with the `BaseApp`.

```go
func (app *BaseApp) FinalizeBlock(req abci.RequestFinalizeBlock) abci.ResponseFinalizeBlock {

    var abciRes abci.ResponseFinalizeBlock
    defer func() {
        // call the streaming service hook with the FinalizeBlock messages
        for _, abciListener := range app.abciListeners {
            ctx := app.finalizeState.ctx
            blockHeight := ctx.BlockHeight()
            if app.abciListenersAsync {
                go func(req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) {
                    if err := app.abciListener.FinalizeBlock(blockHeight, req, res); err != nil {
                        app.logger.Error("FinalizeBlock listening hook failed", "height", blockHeight, "err", err)
                    }
                }(req, abciRes)
            } else {
                if err := app.abciListener.ListenFinalizeBlock(blockHeight, req, res); err != nil {
                    app.logger.Error("FinalizeBlock listening hook failed", "height", blockHeight, "err", err)
                    if app.stopNodeOnABCIListenerErr {
                        os.Exit(1)
                    }
                }
            }
        }
    }()

    ...

    return abciRes
}
```

```go
func (app *BaseApp) Commit() abci.ResponseCommit {

    ...

    res := abci.ResponseCommit{
        Data:         commitID.Hash,
        RetainHeight: retainHeight,
    }

    // call the streaming service hook with the Commit messages
    for _, abciListener := range app.abciListeners {
        ctx := app.deliverState.ctx
        blockHeight := ctx.BlockHeight()
        changeSet := app.cms.PopStateCache()
        if app.abciListenersAsync {
            go func(res abci.ResponseCommit, changeSet []store.StoreKVPair) {
                if err := app.abciListener.ListenCommit(ctx, res, changeSet); err != nil {
                    app.logger.Error("ListenCommit listening hook failed", "height", blockHeight, "err", err)
                }
            }(res, changeSet)
        } else {
            if err := app.abciListener.ListenCommit(ctx, res, changeSet); err != nil {
                app.logger.Error("ListenCommit listening hook failed", "height", blockHeight, "err", err)
                if app.stopNodeOnABCIListenerErr {
                    os.Exit(1)
                }
            }
        }
    }

    ...

    return res
}
```

#### Go Plugin System

We propose a plugin architecture to load and run `Streaming` plugins and other types of implementations. We will introduce a plugin
system over gRPC that is used to load and run Cosmos-SDK plugins. The plugin system uses [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin).
Each plugin must have a struct that implements the `plugin.Plugin` interface and an `Impl` interface for processing messages over gRPC.
Each plugin must also have a message protocol defined for the gRPC service:

```go
// streaming/plugins/abci/{plugin_version}/interface.go

// Handshake is a common handshake that is shared by streaming and host.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var Handshake = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "ABCI_LISTENER_PLUGIN",
    MagicCookieValue: "ef78114d-7bdf-411c-868f-347c99a78345",
}

// ListenerPlugin is the base struct for all kinds of go-plugin implementations
// It will be included in interfaces of different Plugins
type ABCIListenerPlugin struct {
    // GRPCPlugin must still implement the Plugin interface
    plugin.Plugin
    // Concrete implementation, written in Go. This is only used for plugins
    // that are written in Go.
    Impl baseapp.ABCIListener
}

#### Auxiliary streaming service

The `plugin.Plugin` interface has two methods `Client` and `Server`. For our GRPC service these are `GRPCClient` and `GRPCServer`
The `Impl` field holds the concrete implementation of our `baseapp.ABCIListener` interface written in Go.
Note: this is only used for plugin implementations written in Go.

The advantage of having such a plugin system is that within each plugin authors can define the message protocol in a way that fits their use case.
For example, when state change listening is desired, the `ABCIListener` message protocol can be defined as below (*for illustrative purposes only*).
When state change listening is not desired than `ListenCommit` can be omitted from the protocol.

```protobuf
syntax = "proto3";

...

message Empty {}

message ListenFinalizeBlockRequest {
  RequestFinalizeBlock  req = 1;
  ResponseFinalizeBlock res = 2;
}
message ListenCommitRequest {
  int64                block_height = 1;
  ResponseCommit       res          = 2;
  repeated StoreKVPair changeSet    = 3;
}

// plugin that listens to state changes
service ABCIListenerService {
  rpc ListenFinalizeBlock(ListenFinalizeBlockRequest) returns (Empty);
  rpc ListenCommit(ListenCommitRequest) returns (Empty);
}
```

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

Implementing the service above:

```go
// streaming/plugins/abci/{plugin_version}/grpc.go

var (
    _ baseapp.ABCIListener = (*GRPCClient)(nil)
)

// GRPCClient is an implementation of the ABCIListener and ABCIListenerPlugin interfaces that talks over RPC.
type GRPCClient struct {
    client ABCIListenerServiceClient
}

func (m *GRPCClient) ListenFinalizeBlock(goCtx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
    ctx := sdk.UnwrapSDKContext(goCtx)
    _, err := m.client.ListenDeliverTx(ctx, &ListenDeliverTxRequest{BlockHeight: ctx.BlockHeight(), Req: req, Res: res})
    return err
}

func (m *GRPCClient) ListenCommit(goCtx context.Context, res abci.ResponseCommit, changeSet []store.StoreKVPair) error {
    ctx := sdk.UnwrapSDKContext(goCtx)
    _, err := m.client.ListenCommit(ctx, &ListenCommitRequest{BlockHeight: ctx.BlockHeight(), Res: res, ChangeSet: changeSet})
    return err
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
    // This is the real implementation
    Impl baseapp.ABCIListener
}

func (m *GRPCServer) ListenFinalizeBlock(ctx context.Context, req *ListenFinalizeBlockRequest) (*Empty, error) {
    return &Empty{}, m.Impl.ListenFinalizeBlock(ctx, req.Req, req.Res)
}

func (m *GRPCServer) ListenCommit(ctx context.Context, req *ListenCommitRequest) (*Empty, error) {
    return &Empty{}, m.Impl.ListenCommit(ctx, req.Res, req.ChangeSet)
}

```

And the pre-compiled Go plugin `Impl`(*this is only used for plugins that are written in Go*):

```go
// streaming/plugins/abci/{plugin_version}/impl/plugin.go

// Plugins are pre-compiled and loaded by the plugin system

// ABCIListener is the implementation of the baseapp.ABCIListener interface
type ABCIListener struct{}

func (m *ABCIListenerPlugin) ListenFinalizeBlock(ctx context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
    // send data to external system
}

func (m *ABCIListenerPlugin) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []store.StoreKVPair) error {
    // send data to external system
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: grpc_abci_v1.Handshake,
        Plugins: map[string]plugin.Plugin{
           "grpc_plugin_v1": &grpc_abci_v1.ABCIListenerGRPCPlugin{Impl: &ABCIListenerPlugin{}},
        },

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

We will introduce a plugin loading system that will return `(interface{}, error)`.
This provides the advantage of using versioned plugins where the plugin interface and gRPC protocol change over time.
In addition, it allows for building independent plugin that can expose different parts of the system over gRPC.

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

The `NewStreamingPlugin` and `RegisterStreamingPlugin` functions are used to register a plugin with the App's BaseApp.

e.g. in `NewSimApp`:

```go
func NewSimApp(
    logger log.Logger,
    db corestore.KVStoreWithBatch,
    traceStore io.Writer,
    loadLatest bool,
    appOpts servertypes.AppOptions,
    baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {

    ...

    keys := sdk.NewKVStoreKeys(
       authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
       minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
       govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
       evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
    )

    ...

    // register streaming services
    streamingCfg := cast.ToStringMap(appOpts.Get(baseapp.StreamingTomlKey))
    for service := range streamingCfg {
        pluginKey := fmt.Sprintf("%s.%s.%s", baseapp.StreamingTomlKey, service, baseapp.StreamingPluginTomlKey)
        pluginName := strings.TrimSpace(cast.ToString(appOpts.Get(pluginKey)))
        if len(pluginName) > 0 {
            logLevel := cast.ToString(appOpts.Get(flags.FlagLogLevel))
            plugin, err := streaming.NewStreamingPlugin(pluginName, logLevel)
            if err != nil {
                tmos.Exit(err.Error())
            }
            if err := baseapp.RegisterStreamingPlugin(bApp, appOpts, keys, plugin); err != nil {
                tmos.Exit(err.Error())
            }
        }
    }

    return app
```

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

The plugin system will be configured within an App's TOML configuration files.

```toml
# gRPC streaming
[streaming]

# ABCI streaming service
[streaming.abci]

# The plugin version to use for ABCI listening
plugin = "abci_v1"

# List of kv store keys to listen to for state changes.
# Set to ["*"] to expose all keys.
keys = ["*"]

# Enable abciListeners to run asynchronously.
# When abciListenersAsync=false and stopNodeOnABCIListenerErr=false listeners will run synchronized but will not stop the node.
# When abciListenersAsync=true stopNodeOnABCIListenerErr will be ignored.
async = false

# Whether to stop the node on message deliver error.
stop-node-on-err = true
```

There will be four parameters for configuring `ABCIListener` plugin: `streaming.abci.plugin`, `streaming.abci.keys`, `streaming.abci.async` and `streaming.abci.stop-node-on-err`.
`streaming.abci.plugin` is the name of the plugin we want to use for streaming, `streaming.abci.keys` is a set of store keys for stores it listens to,
`streaming.abci.async` is bool enabling asynchronous listening and `streaming.abci.stop-node-on-err` is a bool that stops the node when true and when operating
on synchronized mode `streaming.abci.async=false`. Note that `streaming.abci.stop-node-on-err=true` will be ignored if `streaming.abci.async=true`.

The configuration above support additional streaming plugins by adding the plugin to the `[streaming]` configuration section
and registering the plugin with `RegisterStreamingPlugin` helper function.

Note the that each plugin must include `streaming.{service}.plugin` property as it is a requirement for doing the lookup and registration of the plugin
with the App. All other properties are unique to the individual services.

#### Encoding and decoding streams

ADR-038 introduces the interfaces and types for streaming state changes out from KVStores, associating this
data with their related ABCI requests and responses, and registering a service for consuming this data and streaming it to some destination in a final format.
Instead of prescribing a final data format in this ADR, it is left to a specific plugin implementation to define and document this format.
We take this approach because flexibility in the final format is necessary to support a wide range of streaming service plugins. For example,
the data format for a streaming service that writes the data out to a set of files will differ from the data format that is written to a Kafka topic.

## Consequences

These changes will provide a means of subscribing to KVStore state changes in real time.

### Backwards Compatibility

* This ADR changes the `CommitMultiStore` interface, implementations supporting the previous version of this interface will not support the new one

### Positive

* Ability to listen to KVStore state changes in real time and expose these events to external consumers

### Negative

* Changes `CommitMultiStore` interface and its implementations

### Neutral

* Introduces additional- but optional- complexity to configuring and running a cosmos application
* If an application developer opts to use these features to expose data, they need to be aware of the ramifications/risks of that data exposure as it pertains to the specifics of their application
