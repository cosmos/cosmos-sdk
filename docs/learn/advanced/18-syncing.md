---
sidebar_position: 1
---

# Syncing

Syncing is the process of downloading the blockchain and state from a remote node.

There are the following types of syncing:  

1. Genesis Sync: operator downloads genesis, sets peers and syncs the blockchain by executing all blocks since genesis.    
   > **NOTE:** requires a peer node that doesn't prune blocks (CometBFT state).

2. State Sync: operator need to set `trust_height` and `trust_hash` of a block that s/he trusts as well as `trust_period` - the time how long the operator will trust the given block height. The app will query peer nodes to stream state from that blocks as well as all following blocks. Once the state is downloaded, the node will start executing all blocks since the `trust_height`.  
   > **NOTE:** requires a peer node that provides a state sync and is trusted by the validator. 

3. Snapshot Sync: operator downloads a snapshot from snapshot provider (usually validators) s/he trusts. From there the operator needs to unpack the snapshot to the node directory, update config, set peers, set moniker, setup validator keys. Once the data directory is properly configured, the validator can start a node. The node will query blocks since the snapshot and execute them.  
   > **NOTE:** requires a snapshot directory that is trusted by the validator.  
   
Genesis Sync has zero trust assumptions, but it's the most resource heavy. It also requires a node with all blocks - many validators prune the blocks to optimize space.  



## Observing syncing progress

> ### Note: This section applies to comet users
>
> To enable the Prometheus metrics, set `instrumentation.prometheus=true` in your config file. Metrics will be served under `/metrics` on `26660` port by default. Listen address can be changed in the config file (see `instrumentation.prometheus_listen_addr`).
>
> More [here](https://github.com/cometbft/cometbft/blob/main/docs/explanation/core/metrics.md).

### Block Sync Metrics

They are defined [here](https://github.com/cometbft/cometbft/blob/main/internal/blocksync/metrics.go) and are accessible from the node's metrics endpoint.

- `blocksync_syncing`: Indicates whether a node is currently block syncing.
- `blocksync_num_txs`: Number of transactions in the latest block.
- `blocksync_total_txs`: Total number of transactions.
- `blocksync_block_size_bytes`: Size of the latest block in bytes.
- `blocksync_latest_block_height`: Height of the latest block.

### Block sync configuration

```toml
[blocksync]
version = "v0" # version of the block sync protocol to use
```

### State Sync Metrics

They are defined [here](https://github.com/cometbft/cometbft/blob/main/statesync/metrics.go) and are accessible from the node's metrics endpoint.

- `statesync_syncing`: Indicates whether a node is currently state syncing.

### State sync configuration

```toml
[statesync]
enable = true # set to true to enable state sync
rpc_servers = "" # comma-separated list of RPC servers for state sync
trust_height = 0 # block height to trust for state sync
trust_hash = "" # block hash to trust for state sync
trust_period = "168h0m0s" # trust period for light client verification
discovery_time = "15s" # time to spend discovering snapshots before picking one
temp_dir = "" # directory for temporary state sync files
chunk_request_timeout = "10s" # timeout for chunk requests
chunk_fetchers = "4" # number of concurrent chunk fetchers
```

### Checking if sync is complete

Query the status using the app cli:

```bash
app q consensus comet syncing
```

Query for the node status using the REST or GRPC API:  

REST example:  
```bash  
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/syncing  
```  

Expected response:  
```json  
{  
  "syncing": false  
}  
```  
  
GRPC example:  
```bash  
grpcurl -plaintext localhost:9090 cosmos.base.tendermint.v1beta1.Service/GetSyncing  
```  

The response includes `SyncInfo.CatchingUp` field  
Syncing is complete when this field is `false`  




