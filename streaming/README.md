# Cosmos-SDK Plugins

This package contains an extensible plugin system for the Cosmos-SDK. The plugin system leverages the [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) system. This system is designed to work over RCP. 

Although the `go-plugin` is built to work over RCP, it is currently only designed to work over a local network.

## Pre requisites
For an overview of supported features by the `go-plugin` system, please see https://github.com/hashicorp/go-plugin. The `go-plugin` documentation is located [here](https://github.com/hashicorp/go-plugin/tree/master/docs). You can also directly visit any of the links below:
- [Writing plugins without Go](https://github.com/hashicorp/go-plugin/blob/master/docs/guide-plugin-write-non-go.md) 
- [Go Plugin Tutorial](https://github.com/hashicorp/go-plugin/blob/master/docs/extensive-go-plugin-tutorial.md)
- [Plugin Internals](https://github.com/hashicorp/go-plugin/blob/master/docs/internals.md)
- [Plugin Architecture](https://www.youtube.com/watch?v=SRvm3zQQc1Q) (start here)

## Exposing plugins

To expose plugins to the plugin system, you will need to:
1. Implement the gRPC message protocol service of the plugin
2. Build the plugin binary
3. Export it

The plugin system will look for the binary in the `env` variable `COSMOS_SDK_{PLUGIN_NAME}` above and if it does not find it, it will error out.
Note, the plugin in is the UPPERCASE name of the `streaming.plugin` TOML configuration setting.

Example:
```shell
export COSMOS_SDK_GRPC_ABCI_V1=.../cosmos-sdk/streaming/plugins/abci/grpc_abci_v1/examples/plugin-go/stdout
```

```toml
# gRPC streaming
[streaming]

# Turn on/off gRPC streaming
enable = true

# List of kv store keys to stream out via gRPC
# Set to ["*"] to expose all keys.
keys = ["*"]

# The plugin name used for streaming via gRPC
plugin = "grpc_abci_v1"

# Stop node on deliver error.
# When false, the node will operate in a fire-and-forget mode
# When true, the node will panic with an error.
stop-node-on-err = true
```

## Streaming plugins

List of support streaming plugins

- [ABCI State Streaming Plugin](./plugins/abci/README.md)
