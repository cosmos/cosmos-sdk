<!--
order: 7
-->

# gRPC and REST, and Other Endpoints

This document prevents an overview of all the endpoints a node exposes: gRPC, REST as well as some other endpoints. {synopsis}

## An Overview of All Endpoints

Each node exposes the following endpoints for users to interact with a node, each endpoint is served on a different port. Details on how to configure each endpoint is provided in the endpoint's own section.

- the gRPC server (default port: `9090`),
- the REST server (default port: `1317`),
- the Tendermint RPC endpoint (default port: `26657`).

::: tip
The node also exposes some other endpoints, such as the Tendermint P2P endpoint, or the [Prometheus endpoint](https://docs.tendermint.com/master/nodes/metrics.html#metrics), which are not directly related to the Cosmos SDK. Please refer to the [Tendermint documentation](https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#configuration) for more information.
:::

## gRPC Server

Cosmos SDK v0.40 introduced Protobuf as the main [encoding](./encoding) library, and this brings a wide range of Protobuf-based tools that can be plugged into the SDK. One such tool is [gRPC](https://grpc.io), a modern open source high performance RPC framework that has decent client support in several languages.

Each module exposes [`Msg` and `Query` Protobuf services](../building-modules/messages-and-queries.md) to define state transitions and state queries. These services are hooked up to gRPC via the following function inside the application:

https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc4/server/types/app.go#L39-L41

The `grpc.Server` is a concrete gRPC server, which spawns and serves any gRPC requests. This server can be configured inside `app.toml`:

- the `grpc.enable = true|false` field defines if the gRPC server should be enabled. Defaults to `true`.
- the `grpc.address = {string}` field defines the address (really, the port, since the host should be kept at `0.0.0.0`) the server should bind to. Defaults to `0.0.0.0:9000`.

Once the gRPC server is started, you can send requests using a gRPC client. Some examples are given in our [Interact with the Node](../run-node/interact-node.md#using-grpc) tutorial.

## REST Server

In Cosmos SDK v0.40, the node continues to serve a REST server. However, the existing routes present in version v0.39 and earlier are now marked as deprecated, and new routes have been added via gRPC-gateway.

### gRPC-gateway REST Routes

[gRPC-gateway](https://grpc-ecosystem.github.io/grpc-gateway/) is a tool to transform gRPC endpoints into REST endpoints. For each RPC endpoint defined in a Protobuf service, the SDK offers a REST equivalent. For instances, querying a balance could be done via the `/cosmos.bank.v1beta1.Query/AllBalances` gRPC endpoint, or via the gRPC-gateway `"/cosmos/bank/v1beta1/balances/{address}"` REST endpoint:

## Next {hide}

Learn about [events](./events.md) {hide}
