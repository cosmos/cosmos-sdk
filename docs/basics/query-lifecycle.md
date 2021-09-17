<!--
order: 3
-->

# Query Lifecycle

This document describes the lifecycle of a query in a SDK application, from the user interface to application stores and back. {synopsis}

## Pre-requisite Readings

- [Transaction Lifecycle](./tx-lifecycle.md) {prereq}

## Query Creation

A [**query**](../building-modules/messages-and-queries.md#queries) is a request for information made by end-users of applications through an interface and processed by a full-node. Users can query information about the network, the application itself, and application state directly from the application's stores or modules. Note that queries are different from [transactions](../core/transactions.md) (view the lifecycle [here](./tx-lifecycle.md)), particularly in that they do not require consensus to be processed (as they do not trigger state-transitions); they can be fully handled by one full-node.

For the purpose of explaining the query lifecycle, let's say `MyQuery` is requesting a list of delegations made by a certain delegator address in the application called `simapp`. As to be expected, the [`staking`](../../x/staking/spec/README.md) module handles this query. But first, there are a few ways `MyQuery` can be created by users.

### CLI

The main interface for an application is the command-line interface. Users connect to a full-node and run the CLI directly from their machines - the CLI interacts directly with the full-node. To create `MyQuery` from their terminal, users type the following command:

```bash
simd query staking delegations <delegatorAddress>
```

This query command was defined by the [`staking`](../../x/staking/spec/README.md) module developer and added to the list of subcommands by the application developer when creating the CLI.

Note that the general format is as follows:

```bash
simd query [moduleName] [command] <arguments> --flag <flagArg>
```

To provide values such as `--node` (the full-node the CLI connects to), the user can use the [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml) config file to set them or provide them as flags.

The CLI understands a specific set of commands, defined in a hierarchical structure by the application developer: from the [root command](../core/cli.md#root-command) (`simd`), the type of command (`Myquery`), the module that contains the command (`staking`), and command itself (`delegations`). Thus, the CLI knows exactly which module handles this command and directly passes the call there.

### gRPC

::: warning
A patch introduced in `go-grpc v1.34.0` made gRPC incompatible with the `gogoproto` library, making some [gRPC queries](https://github.com/cosmos/cosmos-sdk/issues/8426) panic. As such, the SDK requires that `go-grpc <=v1.33.2` is installed in your `go.mod`.

To make sure that gRPC is working properly, it is **highly recommended** to add the following line in your application's `go.mod`:

```
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2
```

Please see [issue #8392](https://github.com/cosmos/cosmos-sdk/issues/8392) for more info.
:::

Another interface through which users can make queries, introduced in Cosmos SDK v0.40, is [gRPC](https://grpc.io) requests to a [gRPC server](../core/grpc_rest.md#grpc-server). The endpoints are defined as [Protocol Buffers](https://developers.google.com/protocol-buffers) service methods inside `.proto` files, written in Protobuf's own language-agnostic interface definition language (IDL). The Protobuf ecosystem developed tools for code-generation from `*.proto` files into various languages. These tools allow to build gRPC clients easily.

One such tool is [grpcurl](https://github.com/fullstorydev/grpcurl), and a gRPC request for `MyQuery` using this client looks like:

```bash
grpcurl \
    -plaintext                                           # We want results in plain test
    -import-path ./proto \                               # Import these .proto files
    -proto ./proto/cosmos/staking/v1beta1/query.proto \  # Look into this .proto file for the Query protobuf service
    -d '{"address":"$MY_DELEGATOR"}' \                   # Query arguments
    localhost:9090 \                                     # gRPC server endpoint
    cosmos.staking.v1beta1.Query/Delegations             # Fully-qualified service method name
```

### REST

Another interface through which users can make queries is through HTTP Requests to a [REST server](../core/grpc_rest.md#rest-server). The REST server is fully auto-generated from Protobuf services, using [gRPC-gateway](https://github.com/grpc-ecosystem/grpc-gateway).

An example HTTP request for `MyQuery` looks like:

```bash
GET http://localhost:1317/cosmos/staking/v1beta1/delegators/{delegatorAddr}/delegations
```

## How Queries are Handled by the CLI

The examples above show how an external user can interact with a node by querying its state. To understand more in details the exact lifecycle of a query, let's dig into how the CLI prepares the query, and how the node handles it. The interactions from the users' perspective are a bit different, but the underlying functions are almost identical because they are implementations of the same command defined by the module developer. This step of processing happens within the CLI, gRPC or REST server and heavily involves a `client.Context`.

### Context

The first thing that is created in the execution of a CLI command is a `client.Context`. A `client.Context` is an object that stores all the data needed to process a request on the user side. In particular, a `client.Context` stores the following:

- **Codec**: The [encoder/decoder](../core/encoding.md) used by the application, used to marshal the parameters and query before making the Tendermint RPC request and unmarshal the returned response into a JSON object. The default codec used by the CLI is Protobuf.
- **Account Decoder**: The account decoder from the [`auth`](../..//x/auth/spec/README.md) module, which translates `[]byte`s into accounts.
- **RPC Client**: The Tendermint RPC Client, or node, to which the request will be relayed to.
- **Keyring**: A [Key Manager](../basics/accounts.md#keyring) used to sign transactions and handle other operations with keys.
- **Output Writer**: A [Writer](https://golang.org/pkg/io/#Writer) used to output the response.
- **Configurations**: The flags configured by the user for this command, including `--height`, specifying the height of the blockchain to query and `--indent`, which indicates to add an indent to the JSON response.

The `client.Context` also contains various functions such as `Query()` which retrieves the RPC Client and makes an ABCI call to relay a query to a full-node.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/client/context.go#L20-L50

The `client.Context`'s primary role is to store data used during interactions with the end-user and provide methods to interact with this data - it is used before and after the query is processed by the full-node. Specifically, in handling `MyQuery`, the `client.Context` is utilized to encode the query parameters, retrieve the full-node, and write the output. Prior to being relayed to a full-node, the query needs to be encoded into a `[]byte` form, as full-nodes are application-agnostic and do not understand specific types. The full-node (RPC Client) itself is retrieved using the `client.Context`, which knows which node the user CLI is connected to. The query is relayed to this full-node to be processed. Finally, the `client.Context` contains a `Writer` to write output when the response is returned. These steps are further described in later sections.

### Arguments and Route Creation

At this point in the lifecycle, the user has created a CLI command with all of the data they wish to include in their query. A `client.Context` exists to assist in the rest of the `MyQuery`'s journey. Now, the next step is to parse the command or request, extract the arguments, and encode everything. These steps all happen on the user side within the interface they are interacting with.

#### Encoding

In our case (querying an address's delegations), `MyQuery` contains an [address](./accounts.md#addresses) `delegatorAddress` as its only argument. However, the request can only contain `[]byte`s, as it will be relayed to a consensus engine (e.g. Tendermint Core) of a full-node that has no inherent knowledge of the application types. Thus, the `codec` of `client.Context` is used to marshal the address.

Here is what the code looks like for the CLI command:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/x/staking/client/cli/query.go#L324-L327

#### gRPC Query Client Creation

The SDK leverages code generated from Protobuf services to make queries. The `staking` module's `MyQuery` service generates a `queryClient`, which the CLI will use to make queries. Here is the relevant code:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/x/staking/client/cli/query.go#L318-L342

Under the hood, the `client.Context` has a `Query()` function used to retrieve the pre-configured node and relay a query to it; the function takes the query fully-qualified service method name as path (in our case: `/cosmos.staking.v1beta1.Query/Delegations`), and arguments as parameters. It first retrieves the RPC Client (called the [**node**](../core/node.md)) configured by the user to relay this query to, and creates the `ABCIQueryOptions` (parameters formatted for the ABCI call). The node is then used to make the ABCI call, `ABCIQueryWithOptions()`.

Here is what the code looks like:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/client/query.go#L65-L91

## RPC

With a call to `ABCIQueryWithOptions()`, `MyQuery` is received by a [full-node](../core/encoding.md) which will then process the request. Note that, while the RPC is made to the consensus engine (e.g. Tendermint Core) of a full-node, queries are not part of consensus and will not be broadcasted to the rest of the network, as they do not require anything the network needs to agree upon.

Read more about ABCI Clients and Tendermint RPC in the Tendermint documentation [here](https://tendermint.com/rpc).

## Application Query Handling

When a query is received by the full-node after it has been relayed from the underlying consensus engine, it is now being handled within an environment that understands application-specific types and has a copy of the state. [`baseapp`](../core/baseapp.md) implements the ABCI [`Query()`](../core/baseapp.md#query) function and handles gRPC queries. The query route is parsed, and it it matches the fully-qualified service method name of an existing service method (most likely in one of the modules), then `baseapp` will relay the request to the relevant module.

Apart from gRPC routes, `baseapp` also handles four different types of queries: `app`, `store`, `p2p`, and `custom`. The first three types (`app`, `store`, `p2p`) are purely application-level and thus directly handled by `baseapp` or the stores, but the `custom` query type requires `baseapp` to route the query to a module's [legacy queriers](../building-modules/query-services.md#legacy-queriers). To learn more about these queries, please refer to [this guide](../core/grpc_rest.md#tendermint-rpc).

Since `MyQuery` has a Protobuf fully-qualified service method name from the `staking` module (recall `/cosmos.staking.v1beta1.Query/Delegations`), `baseapp` first parses the path, then uses its own internal `GRPCQueryRouter` to retrieve the corresponding gRPC handler, and routes the query to the module. The gRPC handler is responsible for recognizing this query, retrieving the appropriate values from the application's stores, and returning a response. Read more about query services [here](../building-modules/query-services.md).

Once a result is received from the querier, `baseapp` begins the process of returning a response to the user.

## Response

Since `Query()` is an ABCI function, `baseapp` returns the response as an [`abci.ResponseQuery`](https://tendermint.com/docs/spec/abci/abci.html#messages) type. The `client.Context` `Query()` routine receives the response and.

### CLI Response

The application [`codec`](../core/encoding.md) is used to unmarshal the response to a JSON and the `client.Context` prints the output to the command line, applying any configurations such as the output type (text, JSON or YAML).

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/client/context.go#L248-L283

And that's a wrap! The result of the query is outputted to the console by the CLI.

## Next {hide}

Read more about [accounts](./accounts.md). {hide}
