---
sidebar_position: 1
---

# Configuration


## minimum-gas-prices

The minimum gas prices a validator is willing to accept for processing a transaction. A transaction's fees must meet the minimum of any denomination specified in this config (e.g. 0.25token1;0.0001token2).

## pruning

Determines the pruning strategy to be used. The possible values are default, nothing, everything, and custom. Pruning only handles pruning of application state (application.db). To prune Comet databases please see [Cometbft docs](https://docs.cometbft.com/v0.37/)

default: the last 362880 states are kept, pruning at 10 block intervals.
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node).
everything: 2 latest states will be kept; pruning at 10 block intervals.
custom: allow pruning options to be manually specified through 'pruning-keep-recent', and 'pruning-interval'.

* pruning-keep-recent:Defines the number of recent state snapshots to be kept in case the pruning strategy is set to custom.
* pruning-interval: Defines the block interval at which pruning will occur, in case the pruning strategy is set to custom.

## halt-height

Contains a non-zero block height at which a node will gracefully halt and shutdown that can be used to assist upgrades and testing. Commitment of state will be attempted on the corresponding block.

## halt-time

Contains a non-zero minimum block time (in Unix seconds) at which a node will gracefully halt and shutdown that can be used to assist upgrades and testing. Commitment of state will be attempted on the corresponding block.

## min-retain-blocks

Defines the minimum block height offset from the current block being committed, such that all blocks past this offset are pruned from CometBFT. It is used as part of the process of determining the ResponseCommit.RetainHeight value during ABCI Commit. A value of 0 indicates that no blocks should be pruned.

## inter-block-cache

Enables inter-block caching.

## index-events

Defines the set of events to be indexed in the form {eventType}.{attributeKey}. If empty, all events will be indexed.

## iavl-cache-size

Sets the size of the IAVL tree cache (in number of nodes).

::: Warning
Using this feature will increase ram consumption
:::

## iavl-lazy-loading

Enables or disables the lazy loading of the IAVL store. Default is false. This feature is to be used for archive nodes, allowing them to have a faster start up time. 

## app-db-backend

Defines the database backend type to use for the application and snapshots DBs. An empty string indicates that a fallback will be used. The first fallback is the deprecated compile-time types.DBBackend value. The second fallback (if the types.DBBackend also isn't set), is the db-backend value set in CometBFT's config.toml.

## Telemetry Configuration

* service-name: Prefixed with keys to separate services.
* enabled: Enables the application telemetry functionality. When enabled, an in-memory sink is also enabled by default. Operators may also enable other sinks such as Prometheus.
* enable-hostname: Enables prefixing gauge values with hostname.
* enable-hostname-label: Enables adding hostname to labels.
* enable-service-label: Enables adding service to labels.
* prometheus-retention-time: When positive, enables a Prometheus metrics sink.
* global-labels: Defines a global set of name/value label tuples applied to all metrics emitted using the wrapper functions defined in telemetry package.

## API Configuration

* enable: Defines if the API server should be enabled.
* swagger: Defines if Swagger documentation should
* address: The address on which the API server listens for incoming requests.
* max-open-connections: The maximum number of concurrent open connections to the API server.
* rpc-read-timeout: The maximum time in seconds allowed for the API server to read an RPC request.
* rpc-write-timeout: The maximum time in seconds allowed for the API server to write an RPC response.
* rpc-max-body-bytes: The maximum size in bytes of the request body that the API server will accept.
* enabled-unsafe-cors: A boolean value that indicates whether CORS (Cross-Origin Resource Sharing) should be enabled. This is considered unsafe and should only be used at your own risk.

## gRPC Configuration

* enable: A boolean value that indicates whether the gRPC server should be enabled.
* address: The address on which the gRPC server listens for incoming requests.
* max-recv-msg-size: The maximum size in bytes of a message that the gRPC server can receive.
* max-send-msg-size: The maximum size in bytes of a message that the gRPC server can send.

## gRPC Web Configuration

* enable: A boolean value that indicates whether the gRPC-web should be enabled. Note that gRPC must also be enabled for this configuration to have any effect.

> Note: gRPCWeb is on the same port as the API (Default: 1317)

# State Sync Configuration

* snapshot-interval: The block interval at which local state sync snapshots are taken. Set to 0 to disable.
* snapshot-keep-recent: The number of recent snapshots to keep and serve. Set to 0 to keep all.

## State Streaming Configuration

* keys: A list of KV store keys to stream out via gRPC. The store key names must match the module's StoreKey name. Use ["*"] to expose all keys.
* plugin: The plugin name used for streaming via gRPC. Streaming is only enabled if this is set. Supported plugins: abci.
* stop-node-on-err: A boolean value that indicates whether to stop the node on message delivery error.

## Mempool Configuration

* max-txs: The maximum number of transactions that the mempool can hold. Set to 0 for an unbounded amount, -1 to disable transactions from being inserted into the mempool, or a positive number to limit the number of transactions in the mempool by the specified amount. This configuration only applies to SDK built-in app-side mempool implementations.
