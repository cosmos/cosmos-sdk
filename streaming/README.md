# Cosmos-SDK Plugins

This package contains an extensible plugin system for the Cosmos-SDK. The plugin system leverages the [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin) system. This system is designed to work over RCP. 

Although the `go-plugin` is built to work over RCP, it is currently only designed to work over a local network.

## Pre requisites
For an overview of supported features by the `go-plugin` system, please see https://github.com/hashicorp/go-plugin. The `go-plugin` documentation is located [here](https://github.com/hashicorp/go-plugin/tree/master/docs). You can also directly visit any of the links below:
- [Writing plugins without Go](https://github.com/hashicorp/go-plugin/blob/master/docs/guide-plugin-write-non-go.md) 
- [Go Plugin Tutorial](https://github.com/hashicorp/go-plugin/blob/master/docs/extensive-go-plugin-tutorial.md)
- [Plugin Internals](https://github.com/hashicorp/go-plugin/blob/master/docs/internals.md)
- [Plugin Architecture](https://www.youtube.com/watch?v=SRvm3zQQc1Q) (start here)

## Plugins

Plugins are configured from within an App using the `AppOptions` loaded from the `toml` configuration files. Read each plugin's `README.md` below for further documentation.

- [State Streaming Plugin (ABCI)](./plugins/abci/README.md)