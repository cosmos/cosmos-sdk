<!--
order: 7
-->

# gRPC, REST, and Tendermint Endpoints

This document presents an overview of all the endpoints a node exposes: gRPC, REST as well as some other endpoints. {synopsis}

## An Overview of All Endpoints

Each node exposes the following endpoints for users to interact with a node, each endpoint is served on a different port. Details on how to configure each endpoint is provided in the endpoint's own section.

- the gRPC server (default port: `9090`),
- the REST server (default port: `1317`),
- the Tendermint RPC endpoint (default port: `26657`).

::: tip
The node also exposes some other endpoints, such as the Tendermint P2P endpoint, or the [Prometheus endpoint](https://docs.tendermint.com/master/nodes/metrics.html#metrics), which are not directly related to the Cosmos SDK. Please refer to the [Tendermint documentation](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#configuration) for more information about these endpoints.
:::

## gRPC Server

::: warning
A patch introduced in `go-grpc v1.34.0` made gRPC incompatible with the `gogoproto` library, making some [gRPC queries](https://github.com/cosmos/cosmos-sdk/issues/8426) panic. As such, the SDK requires that `go-grpc <=v1.33.2` is installed in your `go.mod`.

To make sure that gRPC is working properly, it is **highly recommended** to add the following line in your application's `go.mod`:

```
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2
```

Please see [issue #8392](https://github.com/cosmos/cosmos-sdk/issues/8392) for more info.
:::

Cosmos SDK v0.40 introduced Protobuf as the main [encoding](./encoding) library, and this brings a wide range of Protobuf-based tools that can be plugged into the SDK. One such tool is [gRPC](https://grpc.io), a modern open source high performance RPC framework that has decent client support in several languages.

Each module exposes [`Msg` and `Query` Protobuf services](../building-modules/messages-and-queries.md) to define state transitions and state queries. These services are hooked up to gRPC via the following function inside the application:

<https://github.com/cosmos/cosmos-sdk/blob/v0.41.0/server/types/app.go#L39-L41>

The `grpc.Server` is a concrete gRPC server, which spawns and serves any gRPC requests. This server can be configured inside `~/.simapp/config/app.toml`:

- `grpc.enable = true|false` field defines if the gRPC server should be enabled. Defaults to `true`.
- `grpc.address = {string}` field defines the address (really, the port, since the host should be kept at `0.0.0.0`) the server should bind to. Defaults to `0.0.0.0:9090`.

:::tip
`~/.simapp` is the directory where the node's configuration and databases are stored. By default, it's set to `~/.{app_name}`.
:::

Once the gRPC server is started, you can send requests to it using a gRPC client. Some examples are given in our [Interact with the Node](../run-node/interact-node.md#using-grpc) tutorial.

An overview of all available gRPC endpoints shipped with the Cosmos SDK is [Protobuf documention](./proto-docs.md).

## REST Server

In Cosmos SDK v0.40, the node continues to serve a REST server. However, the existing routes present in version v0.39 and earlier are now marked as deprecated, and new routes have been added via gRPC-gateway.

All routes are configured under the following fields in `~/.simapp/config/app.toml`:

- `api.enable = true|false` field defines if the REST server should be enabled. Defaults to `false`.
- `api.address = {string}` field defines the address (really, the port, since the host should be kept at `0.0.0.0`) the server should bind to. Defaults to `tcp://0.0.0.0:1317`.
- some additional API configuration options are defined in `~/.simapp/config/app.toml`, along with comments, please refer to that file directly.

### gRPC-gateway REST Routes

If, for various reasons, you cannot use gRPC (for example, you are building a web application, and browsers don't support HTTP2 on which gRPC is built), then the SDK offers REST routes via gRPC-gateway.

[gRPC-gateway](https://grpc-ecosystem.github.io/grpc-gateway/) is a tool to expose gRPC endpoints as REST endpoints. For each RPC endpoint defined in a Protobuf service, the SDK offers a REST equivalent. For instance, querying a balance could be done via the `/cosmos.bank.v1beta1.Query/AllBalances` gRPC endpoint, or alternatively via the gRPC-gateway `"/cosmos/bank/v1beta1/balances/{address}"` REST endpoint: both will return the same result. For each RPC method defined in a Protobuf service, the corresponding REST endpoint is defined as an option:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.41.0/proto/cosmos/bank/v1beta1/query.proto#L19-L22

For application developers, gRPC-gateway REST routes needs to be wired up to the REST server, this is done by calling the `RegisterGRPCGatewayRoutes` function on the ModuleManager.

### Legacy REST API Routes

The REST routes present in Cosmos SDK v0.39 and earlier are marked as deprecated via a [HTTP deprecation header](https://tools.ietf.org/id/draft-dalal-deprecation-header-01.html). They are still maintained to keep backwards compatibility, but will be removed in v0.41. For updating from Legacy REST routes to new gRPC-gateway REST routes, please refer to our [migration guide](../migrations/rest.md).

For application developers, Legacy REST API routes needs to be wired up to the REST server, this is done by calling the `RegisterRESTRoutes` function on the ModuleManager.

### Swagger

A [Swagger](https://swagger.io/) (or OpenAPIv2) specification file is exposed under the `/swagger` route on the API server. Swagger is an open specification describing the API endpoints a server serves, including description, input arguments, return types and much more about each endpoint.

Enabling the `/swagger` endpoint is configurable inside `~/.simapp/config/app.toml` via the `api.swagger` field, which is set to true by default.

For application developers, you may want to generate your own Swagger definitions based on your custom modules. The SDK's [Swagger generation script](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc4/scripts/protoc-swagger-gen.sh) is a good place to start.

## Tendermint RPC

Independently from the Cosmos SDK, Tendermint also exposes a RPC server. This RPC server can be configured by tuning parameters under the `rpc` table in the `~/.simapp/config/config.toml`, the default listening address is `tcp://0.0.0.0:26657`. An OpenAPI specification of all Tendermint RPC endpoints is available [here](https://docs.tendermint.com/master/rpc/).

Some Tendermint RPC endpoints are directly related to the Cosmos SDK:

- `/abci_query`: this endpoint will query the application for state. As the `path` parameter, you can send the following strings:
  - any Protobuf fully-qualified service method, such as `/cosmos.bank.v1beta1.Query/AllBalances`. The `data` field should then include the method's request parameter(s) encoded as bytes using Protobuf.
  - `/app/simulate`: this will simulate a transaction, and return some information such as gas used.
  - `/app/version`: this will return the application's version.
  - `/store/{path}`: this will query the store directly.
  - `/p2p/filter/addr/{port}`: this will return a filtered list of the node's P2P peers by address port.
  - `/p2p/filter/id/{id}`: this will return a filtered list of the node's P2P peers by ID.
- `/broadcast_tx_{aync,async,commit}`: these 3 endpoint will broadcast a transaction to other peers. CLI, gRPC and REST expose [a way to broadcast transations](./transactions.md#broadcasting-the-transaction), but they all use these 3 Tendermint RPCs under the hood.

## Comparison Table

| Name           | Advantages                                                                                                                                                            | Disadvantages                                                                                                 |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| gRPC           | - can use code-generated stubs in various languages<br>- supports streaming and bidirectional communication (HTTP2)<br>- small wire binary sizes, faster transmission | - based on HTTP2, not available in browsers<br>- learning curve (mostly due to Protobuf)                      |
| REST           | - ubiquitous<br>- client libraries in all languages, faster implementation<br>                                                                                        | - only supports unary request-response communication (HTTP1.1)<br>- bigger over-the-wire message sizes (JSON) |
| Tendermint RPC | - easy to use                                                                                                                                                         | - bigger over-the-wire message sizes (JSON)                                                                   |

## Next {hide}

Learn about [the CLI](./cli.md) {hide}
