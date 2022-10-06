# ABCI and State Streaming Plugin (gRPC)

The `BaseApp` package contains the interface for a `ABCIListener` used to write state changes out from individual KVStores to external systems, as described in [ADR-038](../docs/architecture/adr-038-state-listening.md).

Specific `ABCIListener` implementations are written and loaded as plugins by extending the above interface with a `plugin.GRPCPlugin` interface that adds the `Client` and `Server` methods required by the `go-plugin` system to load the plugin and communicate over gRPC. 

```go
// ListenerGRPCPlugin is the implementation of plugin.GRPCPlugin, so we can serve/consume this.
type ListenerGRPCPlugin struct {
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


The `ABCIListener` gRCP protocol is defined below. See [grpc.go](./grpc_abci_v1/grpc.go) and [interface.go](./grpc_abci_v1/interface.go) for how the `ABCIListener` and `plugin.GRPCPlugin` implementations come together.

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

`ABCIListener` plugins registered with `baseapp.StreamingService` during App startup. The plugin is configured from with an App using the `AppOptions` loaded from the TOML configuration files. Every `StreamingService` will be configured within the `[streaming]` TOML mapping. The plugin TOML configuration is as follows.

```toml
# gRPC streaming service to external systems
[streaming]

# Turn on/off gRPC streaming
enable = true

# List of kv store keys to stream out via gRPC. (Optional)
# Set to ["*"] to expose all keys.
keys = ["*"]

# The name of the plugin used for streaming data over gRPC
plugin = "abci"

# stop the node in case of a delivery error. (true|false)
stop-node-on-err = true
```

## Processing Requests

See the [grpc_abci_v1/examples/](./grpc_abci_v1/examples) for Go and non Go language implementation examples.