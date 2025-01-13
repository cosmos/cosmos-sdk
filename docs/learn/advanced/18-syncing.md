---
sidebar_position: 1
---

# Syncing

Syncing is the process of downloading the blockchain and state from a remote node.

There are two types of syncing:

1. Blockchain syncing.   
    With block sync a node is downloading all of the data of an application from genesis and verifying it. 
2. State syncing.   
    With state sync your node will download data related to the head or near the head of the chain and verify the data. This leads to drastically shorter times for joining a network.


## Observing syncing progress

> ### Note: This section applies to comet users.

### Block Sync Metrics

- `blocksync_syncing`: Indicates whether a node is currently block syncing.
- `blocksync_num_txs`: Number of transactions in the latest block.
- `blocksync_total_txs`: Total number of transactions.
- `blocksync_block_size_bytes`: Size of the latest block in bytes.
- `blocksync_latest_block_height`: Height of the latest block.

### Block sync configuration

```toml
blocksync:
    version: "v0"
```

- `version`: The version of the block sync protocol to use.

### State Sync Metrics

- `statesync_syncing`: Indicates whether a node is currently state syncing.

### State sync configuration

```toml
statesync:
    enable: true
    rpc_servers: ""
    trust_height: 0
    trust_hash: ""
    trust_period: 168h0m0s
    discovery_time: 15s
    temp_dir: ""
    chunk_request_timeout: 10s
    chunk_fetchers: "4"
```

- `enable`: Set to true to enable state sync.
- `rpc_servers`: Comma-separated list of RPC servers for state sync.
- `trust_height`: Block height to trust for state sync.
- `trust_hash`: Block hash to trust for state sync.
- `trust_period`: Trust period for light client verification.
- `discovery_time`: Time to spend discovering snapshots before picking one.
- `temp_dir`: Directory for temporary state sync files.
- `chunk_request_timeout`: Timeout for chunk requests.
- `chunk_fetchers`: Number of concurrent chunk fetchers.

### Checking if sync is complete

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




