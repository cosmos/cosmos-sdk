# File Streaming Service

This pkg contains an implementation of the [StreamingService](../../../baseapp/streaming.go) that writes
the data stream out to files on the local filesystem. This process is performed synchronously with the message processing
of the state machine.

## Configuration

The `file.StreamingService` is configured from within an App using the `AppOptions` loaded from the app.toml file:

```toml
[store]
    streamers = [ # if len(streamers) > 0 we are streaming
        "file", # name of the streaming service, used by constructor
    ]

[streamers]
    [streamers.file]
        keys = ["list", "of", "store", "keys", "we", "want", "to", "expose", "for", "this", "streaming", "service"]
        write_dir = "path to the write directory"
        prefix = "optional prefix to prepend to the generated file names"
```

We turn the service on by adding its name, "file", to `store.streamers`- the list of streaming services for this App to employ.

In `streamers.file` we include three configuration parameters for the file streaming service:

1. `streamers.x.keys` contains the list of `StoreKey` names for the KVStores to expose using this service.
In order to expose *all* KVStores, we can include `*` in this list. An empty list is equivalent to turning the service off.
2. `streamers.file.write_dir` contains the path to the directory to write the files to.
3. `streamers.file.prefix` contains an optional prefix to prepend to the output files to prevent potential collisions
with other App `StreamingService` output files.

### Encoding

For each block, two files are created and names `block-{N}-meta` and `block-{N}-data`, where `N` is the block number.

The meta file contains the protobuf encoded message `BlockMetadata` which contains the abci event requests and responses of the block:

```protobuf
message BlockMetadata {
    message DeliverTx {
        tendermint.abci.RequestDeliverTx request = 1;
        tendermint.abci.ResponseDeliverTx response = 2;
    }
    tendermint.abci.RequestBeginBlock request_begin_block = 1;
    tendermint.abci.ResponseBeginBlock response_begin_block = 2;
    repeated DeliverTx deliver_txs = 3;
    tendermint.abci.RequestEndBlock request_end_block = 4;
    tendermint.abci.ResponseEndBlock response_end_block = 5;
    tendermint.abci.ResponseCommit response_commit = 6;
}
```

The data file contains a series of length-prefixed protobuf encoded `StoreKVPair`s representing `Set` and `Delete` operations within the KVStores during the execution of block.

Both meta and data files are prefixed with the length of the data content for consumer to detect completeness of the file, the length is encoded as 8 bytes with big endianness.

The files are written at abci commit event, by default the error happens don't interuppted consensus state machine, and fsync is not called explicitly, it'll have good performance but have the risk of lossing data in face of system crash.

There are several parameters in streaming service constructor to configure the behaviors:

- `stopNodeOnErr`: Call panic when error happens during commit, which will stop the block from commiting succesfully, so when the block is replayed after issue resolved, the process will run again, and data is eventually consistent.
- `fsync`: Call `file.Sync()` after writing the file data, so we don't lose data in face of system crash.
- `outputMetadata`: If `false`, don't write the block meta file, only the data file is outputted.

The default setting is: `stopNodeOnErr=false, fsync=false, outputMetadata=true`, these parameters are not reflected in the configuration system for now, there'll be newly designed plugin and configuration system soon.

### Decoding

The pseudo-code for decoding is like this:

```python
def decode_meta_file(file):
  bz = file.read(8)
  if len(bz) < 8:
    raise "incomplete file exception"
  size = int.from_bytes(bz, 'big')

  if file.size != size + 8:
    raise "incomplete file exception"

  return decode_protobuf_message(BlockMetadata, file)

def decode_data_file(file):
  bz = file.read(8)
  if len(bz) < 8:
    raise "incomplete file exception"
  size = int.from_bytes(bz, 'big')

  if file.size != size + 8:
    raise "incomplete file exception"

  while not file.eof():
    yield decode_length_prefixed_protobuf_message(StoreKVStore, file)
```
