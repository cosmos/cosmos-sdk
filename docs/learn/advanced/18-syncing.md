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

### Checking if sync is complete

- Query for the node status using the REST or GRPC API
- Check the `SyncInfo.CatchingUp` field
- If the field is `false`, then syncing is complete





