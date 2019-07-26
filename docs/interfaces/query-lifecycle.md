# Query Lifecycle

## Prerequisites

[Introduction to Interfaces](./interfaces-intro.md)

## Synopsis

This document describes SDK interfaces in detail through the lifecycle of a query, from the user interface to application stores and back. The query will be referred to as `query`.

- [Interfaces](#interfaces)
- [CLIContext](#clicontext)
- [Tendermint and ABCI](#tendermint-and-abci)
- [Application Query Handling](#application-query-handling)
- [Response](#response)

## Interfaces

A **query** is a request for information made by users of applications. They can query information about the network, the application itself, and application state directly from the application's stores or modules.

For the purpose of explaining a query lifecycle, let's say `query` is requesting a list of delegations made by a certain delegator address in the application called `app`. There are a few ways `query` can be made on the user side.

### CLI

The main interface for an application is the command-line interface. Users run the CLI directly from their machines and type commands to create queries. To create this query from the command-line, users would type the following command:

```
appcli query staking delegations <delegatorAddress>
```

To provide values such as `--chain-id`, the ID of the blockchain to make the query to, the user must use the `config` command to set them or provide them as flags.

This query command is defined by the module developer and added to the list of subcommands by the application developer when creating the CLI. To see the command itself, click [here]().

### REST

Another interface through which users can make queries is through HTTP Requests to a REST server. The REST server contains, among other things, a [`CLIContext`](#clicontext) and [mux](./rest.md#gorilla-mux) router. The request looks like this:

```bash
http://localhost:{PORT}/staking/delegators/{delegatorAddr}/delegations
```

To provide values such as `--chain-id`, the ID of the blockchain to make the query to, the user must configure their REST server with the values or provide them in the request body.

The router automatically routes the `query` HTTP request to the staking module `delegatorDelegationsHandlerFn()` function (to see the handler itself, click [here]()). Since this function is defined within the module and thus has no inherent knowledge of the application `query` belongs to, it takes in the application `codec` and `CLIContext` as parameters.

To read about how the router is used, click [here](./rest.md).

## Request and Command Handling

The interactions from the users' perspective are a bit different, but the underlying functions are almost identical. This section describes how the CLI command or HTTP request is processed, up until the ABCI request is sent. This step of processing heavily involves a `CLIContext`.

### CLIContext

The first thing that is created in the execution of a CLI command is a `CLIContext`, while the REST Server directly provides a `CLIContext` for the REST Request handler. A [Context](../core/context.md) is an immutable object that stores all the data needed to process a request. In particular, a `CLIContext` stores the following:

* **Codec**: The encoder/decoder used by the application, used to marshal the parameters and query before making the Tendermint RPC request and unmarshal the returned response into a JSON object.
* **Account Decoder**: The account decoder from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) module, which translates `[]byte`s into accounts.
* **RPC Client**: The [Tendermint RPC Client](https://github.com/tendermint/tendermint/blob/master/rpc/client/interface.go).
* **Keybase**: A [Key Manager](.//core/accounts-keys.md) used to sign transactions and handle other operations with keys.
* **Output Writer**: A [Writer](https://golang.org/pkg/io/#Writer) used to output the response.
* **Configurations**: The flags configured by the user for this command, including `--height`, specifying the height of the blockchain to query and `--indent`, which indicates to add an indent to the JSON response.

For full specification of the `CLIContext` type, click [here](https://github.com/cosmos/cosmos-sdk/blob/73e5ef7c13c420f9ee879fdf1b60cf0bdc8f325e/client/context/context.go#L36-L59).

### Parameters and Route Creation

The next step is to parse the command or request, extract the arguments, create a `queryRoute`, and encode everything.

In this case, `query` contains a `delegatorAddress` as its only argument. However, the request can only contain `[]byte`s, as it will be relayed to a consensus engine node that has no inherent knowledge of the application types. Thus, the `CLIContext` `codec` is used to marshal the address as the type `QueryDelegatorParams`. All query arguments (e.g. the `staking` module also has `QueryValidatorParams` and `QueryBondsParams`) have their own types that the application `codec` understands how to encode and decode. The module [`querier`](.//building-modules/querier.md) declares these types and the application registers the `codec`s.

A `route` is also created for `query` so that the application will understand which module to route the query to. Baseapp will understand this query to be a `custom` query in the module `staking` with the type `QueryDelegatorDelegations`. Thus, the route will be `"custom/staking/delegatorDelegations"`.

### ABCI Query

The `CLIContext`'s main `query` function takes the `route`, which is now called `path`, and arguments, now called `key`. It first retrieves the RPC Client (called the **node**) configured by the user to relay this query to, and creates the `ABCIQueryOptions` (parameters formatted for the ABCI call). The node is then used to make the ABCI call, `ABCIQueryWithOptions`.

## Tendermint and ABCI

With a call to `ABCIQueryWithOptions()`, `query` arrives at the consensus engine portion of its lifecycle. Nodes running the consensus engine (e.g. Tendermint Core) make ABCI calls to interact with the application. At this point, `query` exists as an ABCI `RequestQuery` and the [ABCI Client](https://github.com/tendermint/tendermint/blob/51b3428f5c0f4fdd2e469147cd90353faa4bd704/abci/client/client.go#L16-L50) calls the ABCI method [`Query()`](https://tendermint.com/docs/spec/abci/abci.html#query) on the application.

Read more about ABCI Clients and Tendermint RPC in the Tendermint documentation [here](https://tendermint.com/rpc).

## Application Query Handling

[Baseapp](../core/baseapp.md) implements the ABCI [`Query()`](../core/baseapp.md#query) function and handles four different types of queries: `app`, `store`, `p2p`, and `custom`. The `queryRoute` is parsed such that the first string must be one of the four options, then the rest of the path is parsed within the subroutines handling each type of query. The first three types are application-level and thus directly handled by Baseapp or the stores, but the `custom` query type requires Baseapp to route the query to a module's [querier](../building-modules/querier.md).

Since `query` is a custom query type, Baseapp first parses the path, then uses the `QueryRouter` to retrieve the corresponding querier.

## Response

Since `Query()` is an ABCI function, Baseapp returns the `query` response as an `abci.ResponseQuery`. The `CLIContext` `query` routine receives the response and, if `--trust-node` is toggled to `false` and a proof needs to be verified, the response is verified with the `CLIContext` `verifyProof` function before the response is returned.

### CLI Response

The application `codec` is used to unmarshal the response to a JSON and the `CLIContext` prints the output to the command line, applying any configurations such as `--indent`.

### REST Response

The REST server uses the `CLIContext` to format the response properly, then uses the HTTP package to write the appropriate response or error.

## Next

Read about how to build a [Command-Line Interface](./cli.md), or a [REST Interface](./rest.md).
