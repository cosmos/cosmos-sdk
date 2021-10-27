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
