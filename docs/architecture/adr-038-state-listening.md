# ADR 038: KVStore state listening

## Changelog

* 11/23/2020: Initial draft

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
We will introduce a plugin system for configuring and running streaming services that write these state changes and their surrounding ABCI message context to different destinations.

### Listening interface

In a new file, `store/types/listening.go`, we will create a `WriteListener` interface for streaming out state changes from a KVStore.

```go
// WriteListener interface for streaming data out from a listenkv.Store
type WriteListener interface {
	// if value is nil then it was deleted
	// storeKey indicates the source KVStore, to facilitate using the same WriteListener across separate KVStores
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

// onWrite writes a KVStore operation to all the WriteListeners
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

#### Streaming Service

We will introduce a new `StreamingService` struct for exposing `WriteListener` data streams to external consumers.
In addition to streaming state changes as `StoreKVPair`s, the struct satisfies an `ABCIListener` interface that plugs
into the BaseApp and relays ABCI requests and responses so that the service can group the state changes with the ABCI
requests that affected them and the ABCI responses they affected. The `ABCIListener` interface also exposes a
`ListenKVStorePair` method which is used by the `StoreKVPairWriteListener` to stream state change events.

```go
// ABCIListener interface used to hook into the ABCI message processing of the BaseApp
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(blockHeight int64, req []byte, res []byte) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(blockHeight int64, req []byte, res []byte) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(blockHeight int64, req []byte, res []byte) error
	// ListenStoreKVPair updates the steaming service with the latest StoreKVPair messages
	ListenStoreKVPair(blockHeight int64, data []byte) error
}

// StreamingService struct for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService struct {
	// Listeners returns the streaming service's listeners for the BaseApp to register
	Listeners map[types.StoreKey][]store.WriteListener
	// ABCIListener interface for hooking into the ABCI messages from inside the BaseApp
	ABCIListener ABCIListener
	// StopNodeOnErr stops the node when true
	StopNodeOnErr bool
}
```

We will also introduce a new `KVStoreWriter` struct as the writer for use in `StoreKVPairWriteListener`. `KVStoreWriter` will expose `ABCIListener` for streaming state change events via `ABCIListener.ListenStoreKVPair`, a `BlockHeight` func for determining the block height for which the state change events belong to and a boolean flag for determine whether the node needs to be stopped or not.

```go
// KVStoreWriter is used so that we do not need to update the underlying
// io.Writer inside the StoreKVPairWriteListener everytime we begin writing
type KVStoreWriter struct {
	blockHeight   func() int64
	listener      ABCIListener
	stopNodeOnErr bool
}

// NewKVStoreWriter create an instance of a KVStoreWriter that sends StoreKVPair data to listening service
func NewKVStoreWriter(
	listener ABCIListener,
	stopNodeOnErr bool,
	blockHeight func() int64,
) *KVStoreWriter {
	return &KVStoreWriter{
		listener:      listener,
		stopNodeOnErr: stopNodeOnErr,
		blockHeight:   blockHeight,
	}
}

// Write satisfies io.Writer
func (iw *KVStoreWriter) Write(b []byte) (int, error) {
	blockHeight := iw.blockHeight()
	if err := iw.listener.ListenStoreKVPair(blockHeight, b); err != nil {
		if iw.stopNodeOnErr {
			panic(err)
		}
		return 0, err
	}
	return len(b), nil
}
```

We will expose a `RegisterStreamingService` function for the App to register streaming services.
```go
// RegisterStreamingService registers the ABCI streaming service provided by the streaming plugin.
func RegisterStreamingService(
	bApp *BaseApp,
	appOpts servertypes.AppOptions,
	kodec codec.BinaryCodec,
	keys map[string]*types.KVStoreKey,
	streamingService interface{},
) error {
	// type checking
	abciListener, ok := streamingService.(ABCIListener)
	if !ok {
		return fmt.Errorf("failed to register streaming service: failed type check %v", streamingService)
	}

	// expose keys
	keysKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingKeysTomlKey)
	exposeKeysStr := cast.ToStringSlice(appOpts.Get(keysKey))
	exposeStoreKeys := exposeStoreKeys(exposeKeysStr, keys)
	stopNodeOnErrKey := fmt.Sprintf("%s.%s", StreamingTomlKey, StreamingStopNodeOnErrTomlKey)
	stopNodeOnErr := cast.ToBool(appOpts.Get(stopNodeOnErrKey))
	blockHeightFn := func() int64 { return bApp.deliverState.ctx.BlockHeight() }
	writer := NewKVStoreWriter(abciListener, stopNodeOnErr, blockHeightFn)
	listener := types.NewStoreKVPairWriteListener(writer, kodec)
	listeners := make(map[types.StoreKey][]types.WriteListener, len(exposeStoreKeys))
	for _, key := range exposeStoreKeys {
		listeners[key] = append(listeners[key], listener)
	}

	bApp.SetStreamingService(StreamingService{
		Listeners:     listeners,
		ABCIListener:  abciListener,
		StopNodeOnErr: stopNodeOnErr,
	})

	return nil
}

func exposeAll(list []string) bool {
	for _, ele := range list {
		if ele == "*" {
			return true
		}
	}
	return false
}

func exposeStoreKeys(keysStr []string, keys map[string]*types.KVStoreKey) []types.StoreKey {
	var exposeStoreKeys []types.StoreKey
	if exposeAll(keysStr) {
		exposeStoreKeys = make([]types.StoreKey, 0, len(keys))
		for _, storeKey := range keys {
			exposeStoreKeys = append(exposeStoreKeys, storeKey)
		}
	} else {
		exposeStoreKeys = make([]types.StoreKey, 0, len(keysStr))
		for _, keyStr := range keysStr {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
	}

	return exposeStoreKeys
}
```

#### BaseApp Registration

We will add a new method to the `BaseApp` to enable the registration of `StreamingService`s:

```go
// SetStreamingService is used to set a streaming service into the BaseApp hooks and load the listeners into the multistore
func (app *BaseApp) SetStreamingService(s StreamingService) {
	// add the listeners for each StoreKey
	for key, lis := range s.Listeners {
		app.cms.AddListeners(key, lis)
	}
	// register the StreamingService within the BaseApp
	// BaseApp will pass BeginBlock, DeliverTx, and EndBlock requests and responses to the streaming services to update their ABCI context
	app.abciListener = s.ABCIListener
	app.stopNodeOnStreamingErr = s.StopNodeOnErr
}
```

We will also modify the `BeginBlock`, `EndBlock`, and `DeliverTx` methods to pass ABCI requests and responses to any streaming service hooks registered
with the `BaseApp`.

```go
func (app *BaseApp) BeginBlock(req abci.RequestBeginBlock) (res abci.ResponseBeginBlock) {

	...

	// call the streaming service hook with the BeginBlock messages
	if app.abciListener != nil {
		blockHeight := app.deliverState.ctx.BlockHeight()
		if app.stopNodeOnStreamingErr {
			reqBz, err := req.Marshal()
			if err != nil {
				panic(err)
			}
			resBz, err := res.Marshal()
			if err != nil {
				panic(err)
			}
			if err := app.abciListener.ListenBeginBlock(blockHeight, reqBz, resBz); err != nil {
				app.logger.Error("BeginBlock listening hook failed", "height", blockHeight, "err", err)
				panic(err)
			}
		} else {
			ebReq, ebRes := req, res
			go func() {
				reqBz, reqErr := ebReq.Marshal()
				if reqErr != nil {
					app.logger.Error("BeginBlock listening hook failed", "height", blockHeight, "err", reqErr)
				}
				resBz, resErr := ebRes.Marshal()
				if resErr != nil {
					app.logger.Error("BeginBlock listening hook failed", "height", blockHeight, "err", resErr)
				}
				if reqErr == nil && resErr == nil {
					if err := app.abciListener.ListenBeginBlock(blockHeight, reqBz, resBz); err != nil {
						app.logger.Error("BeginBlock listening hook failed", "height", blockHeight, "err", err)
					}
				}
			}()
		}
	}

	return res
}
```

```go
func (app *BaseApp) EndBlock(req abci.RequestEndBlock) (res abci.ResponseEndBlock) {

	...

	// call the streaming service hook with the EndBlock messages
	if app.abciListener != nil {
		blockHeight := app.deliverState.ctx.BlockHeight()
		if app.stopNodeOnStreamingErr {
			reqBz, err := req.Marshal()
			if err != nil {
				panic(err)
			}
			resBz, err := res.Marshal()
			if err != nil {
				panic(err)
			}
			if err := app.abciListener.ListenEndBlock(blockHeight, reqBz, resBz); err != nil {
				app.logger.Error("EndBlock listening hook failed", "height", blockHeight, "err", err)
				panic(err)
			}
		} else {
			ebReq, ebRes := req, res
			go func() {
				reqBz, reqErr := ebReq.Marshal()
				if reqErr != nil {
					app.logger.Error("EndBlock listening hook failed", "height", blockHeight, "err", reqErr)
				}
				resBz, resErr := ebRes.Marshal()
				if resErr != nil {
					app.logger.Error("EndBlock listening hook failed", "height", blockHeight, "err", resErr)
				}
				if reqErr == nil && resErr == nil {
					if err := app.abciListener.ListenEndBlock(blockHeight, reqBz, resBz); err != nil {
						app.logger.Error("EndBlock listening hook failed", "height", blockHeight, "err", err)
					}
				}
			}()
		}
	}

	return res
}
```

```go
func (app *BaseApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {

	var abciRes abci.ResponseDeliverTx
	defer func() {
		// call the streaming service hook with the EndBlock messages
		if app.abciListener != nil {
			blockHeight := app.deliverState.ctx.BlockHeight()
			if app.stopNodeOnStreamingErr {
				reqBz, err := req.Marshal()
				if err != nil {
					panic(err)
				}
				resBz, err := abciRes.Marshal()
				if err != nil {
					panic(err)
				}
				if err := app.abciListener.ListenDeliverTx(blockHeight, reqBz, resBz); err != nil {
					app.logger.Error("DeliverTx listening hook failed", "height", blockHeight, "err", err)
					panic(err)
				}
			} else {
				txReq, txRes := req, abciRes
				go func() {
					reqBz, reqErr := txReq.Marshal()
					if reqErr != nil {
						app.logger.Error("DeliverTx listening hook failed", "height", blockHeight, "err", reqErr)
					}
					resBz, resErr := txRes.Marshal()
					if resErr != nil {
						app.logger.Error("DeliverTx listening hook failed", "height", blockHeight, "err", resErr)
					}
					if reqErr == nil && resErr == nil {
						if err := app.abciListener.ListenDeliverTx(blockHeight, reqBz, resBz); err != nil {
							app.logger.Error("DeliverTx listening hook failed", "height", blockHeight, "err", err)
						}
					}
				}()
			}
		}
	}()

	...

	return abciRes
}
```

#### Go Plugin System

We propose a plugin architecture to load and run `StreamingService` and other types of implementations. We will introduce a plugin
system over gRPC that is used to load and run Cosmos-SDK plugins. The plugin system uses [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin).
Each plugin must have a struct that implements the `plugin.Plugin` interface and an `Impl` interface for processing messages over gRPC.
Each plugin must also have a message protocol defined for the gRPC service:

```go
// ListenerPlugin is the base struc for all kinds of go-plugin implementations
// It will be included in interfaces of different Plugins
type ListenerPlugin interface {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl baseapp.ABCIListener
}

func (p *ListenerGRPCPlugin) GRPCServer(_ *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterABCIListenerServiceServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *ListenerGRPCPlugin) GRPCClient(
	_ context.Context,
	_ *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &GRPCClient{client: NewABCIListenerServiceClient(c)}, nil
}
```

The `plugin.Plugin` interface has two methods `Client` and `Server`. For our GRPC service these are `GRPCClient` and `GRPCServer`
The `Impl` field holds the concrete implementation of our `baseapp.ABCIListener` interface written in Go.
Note: this is only used for plugins that are written in Go.

The `StreamingService` protocol:

```protobuf
syntax = "proto3";

package cosmos.sdk.grpc.abci.v1;

option go_package           = "github.com/cosmos/cosmos-sdk/streaming/plugins/abci/grpc_abci_v1";
option java_multiple_files  = true;
option java_outer_classname = "AbciListenerProto";
option java_package         = "network.cosmos.sdk.grpc.abci.v1";

// PutRequest is used for storing ABCI request and response
// and Store KV data for streaming to external grpc service.
message PutRequest {
  int64 block_height      = 1;
  bytes req               = 2;
  bytes res               = 3;
  bytes store_kv_pair     = 4;
  int64 store_kv_pair_idx = 5;
  int64 tx_idx            = 6;
}

message Empty {}

service ABCIListenerService {
  rpc ListenBeginBlock(PutRequest) returns (Empty);
  rpc ListenEndBlock(PutRequest) returns (Empty);
  rpc ListenDeliverTx(PutRequest) returns (Empty);
  rpc ListenStoreKVPair(PutRequest) returns (Empty);
}
```

For the purposes of this ADR we introduce a single state streaming plugin loading system. This loading system returns `(interface{}, error)`.
This provides the advantage of using versioned plugins where the plugin interface and gRPC protocol change over time.

```go
func NewStreamingPlugin(name string, logLevel string) (interface{}, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Output: hclog.DefaultOutput,
		Level:  toHclogLevel(logLevel),
		Name:   fmt.Sprintf("plugin.%s", name),
	})

	// We're a host. Start by launching the streaming process.
	env := os.Getenv(GetPluginEnvKey(name))
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeMap[name],
		Plugins:         PluginMap,
		Cmd:             exec.Command("sh", "-c", env),
		Logger:          logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request streaming plugin
	return rpcClient.Dispense(name)
}

```

The `RegisterStreamingService` function is used during App construction to register a plugin with the App's BaseApp using the `NewStreamingPlugin` method.

e.g. in `NewSimApp`:

```go
func NewSimApp(
	logger log.Logger,
	db dbm.DB,
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

	// enable streaming
	enableKey := fmt.Sprintf("%s.%s", baseapp.StreamingTomlKey, baseapp.StreamingEnableTomlKey)
	if enable := cast.ToBool(appOpts.Get(enableKey)); enable {
		pluginKey := fmt.Sprintf("%s.%s", baseapp.StreamingTomlKey, baseapp.StreamingPluginTomlKey)
		pluginName := cast.ToString(appOpts.Get(pluginKey))
		logLevel := cast.ToString(appOpts.Get(flags.FlagLogLevel))
		plugin, err := streaming.NewStreamingPlugin(pluginName, logLevel)
		if err != nil {
			tmos.Exit(err.Error())
		}
		if err := baseapp.RegisterStreamingService(bApp, appOpts, appCodec, keys, plugin); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
```

#### Configuration

The plugin system will be configured within an App's TOML configuration files.

```toml
# gRPC streaming
[streaming]

# Turn on/off gRPC streaming
enable = true

# List of kv store keys to stream out via gRPC
# Set to ["*"] to expose all keys.
keys = ["*"]

# The plugin name used for streaming via gRPC
plugin = "my_streaming_plugin"

# Stop node on deliver error.
# When false, the node will operate in a fire-and-forget mode
# When true, the node will panic with an error.
stop-node-on-err = true
```

There will be four parameters for configuring the streaming plugin system: `streaming.enable`, `streaming.keys`, `streaming.plugin` and `streaming.stop-node-on-err`.
`streaming.enable` is a bool that turns on or off the streaming service, `streaming.keys` is a set of store keys for stores it listens to,
`streaming.plugin` is the name of the plugin we want to use for streaming and `streaming.stop-node-on-err` is a bool that stops the node when true and operates in a fire-and-forget mode when false.

#### Encoding and decoding streams

ADR-038 introduces the interfaces and types for streaming state changes out from KVStores, associating this
data with their related ABCI requests and responses, and registering a service for consuming this data and streaming it to some destination in a final format.
Instead of prescribing a final data format in this ADR, it is left to a specific plugin implementation to define and document this format.
We take this approach because flexibility in the final format is necessary to support a wide range of streaming service plugins. For example,
the data format for a streaming service that writes the data out to a set of files will differ from the data format that is written to a Kafka topic.

## Consequences

These changes will provide a means of subscribing to KVStore state changes in real time.

### Backwards Compatibility

* This ADR changes the `MultiStore`, `CacheWrap`, and `CacheWrapper` interfaces, implementations supporting the previous version of these interfaces will not support the new ones

### Positive

* Ability to listen to KVStore state changes in real time and expose these events to external consumers

### Negative

* Changes `MultiStore`, `CacheWrap`, and `CacheWrapper` interfaces

### Neutral

* Introduces additional- but optional- complexity to configuring and running a cosmos application
* If an application developer opts to use these features to expose data, they need to be aware of the ramifications/risks of that data exposure as it pertains to the specifics of their application
