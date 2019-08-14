# Query Lifecycle

## Prerequisites

* [Introduction to Interfaces](./interfaces-intro.md)

## Synopsis

This document describes SDK interfaces in detail through the lifecycle of a query, from the user interface to application stores and back. The query will be referred to as `Query`.

- [Interfaces](#interfaces)
- [Request and Command Handling](#request-and-command-handling)
- [Tendermint and ABCI](#tendermint-and-abci)
- [Application Query Handling](#application-query-handling)
- [Response](#response)

## Interfaces

A [**query**](../building-modules/messages-and-queries.md#queries) is a request for information made by end-users of applications through an interface and processed by a full-node. Users can query information about the network, the application itself, and application state directly from the application's stores or modules. Note that queries are different from [transactions](../core/transactions.md) (view the lifecycle [here](../basics/tx-lifecycle.md)), particularly in that they do not require consensus to be processed; they can be fully handled by one full-node.

For the purpose of explaining a query lifecycle, let's say `Query` is requesting a list of delegations made by a certain delegator address in the application called `app`. As to be expected, the [`staking`](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/staking) module handles this query. But first, there are a few ways `Query` can be created by users.

### CLI

The main interface for an application is the command-line interface. Users connect to a full-node and run the CLI directly from their machines - the CLI interacts directly with the full-node. To create `Query` from their terminal, users type the following command:

```bash
appcli query staking delegations <delegatorAddress>
```

Note that the general format is as follows:

```bash
appcli query [moduleName] [command] <arguments> --flag <flagArg>
```

To provide values such as `--node` (the full-node the CLI connects to), the user must use the `config` command to set them or provide them as flags.

This query command was defined by the [`staking`](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/staking) module developer and added to the list of subcommands by the application developer when creating the CLI. The code for this particular command can be found [here](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/client/cli/query.go#L253-L294).

The CLI understands a specific set of commands, defined in a hierarchical structure by the application developer: from the [root command](./cli.md#root-command) (`appcli`), the type of command (`query`), the module that contains the command  (`staking`), and command itself (`delegations`). Thus, the CLI knows exactly which module handles this command and directly passes the call there.

### REST

Another interface through which users can make queries is through HTTP Requests to a [REST server](./rest.md#rest-server). The REST server contains, among other things, a [`CLIContext`](#clicontext) and [mux](./rest.md#gorilla-mux) router. The request looks like this:

```bash
GET http://localhost:{PORT}/staking/delegators/{delegatorAddr}/delegations
```

To provide values such as `--node` (the full-node the CLI connects to) that are required by [`baseReq`](../building-modules/module-interfaces.md#basereq), the user must configure their local REST server with the values or provide them in the request body.

The router automatically routes the `Query` HTTP request to the staking module `delegatorDelegationsHandlerFn()` function (to see the handler itself, click [here](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/client/rest/query.go#L103-L106)). Since this function is defined within the module and thus has no inherent knowledge of the application `Query` belongs to, it takes in the application `codec` and `CLIContext` as parameters.

To summarize, when users interact with the interfaces, they create a CLI command or HTTP request. `Query` is then created when the command is executed or HTTP request is handled.

## Request and Command Handling

The interactions from the users' perspective are a bit different, but the underlying functions are almost identical because they are implementations of the same command defined by the module developer. This step of processing heavily involves a `CLIContext`.

### CLIContext

The first thing that is created in the execution of a CLI command is a `CLIContext`, while the REST Server directly provides a `CLIContext` for the REST Request handler. A [Context](../core/context.md) is an immutable object that stores all the data needed to process a request. In particular, a `CLIContext` stores the following:

* **Codec**: The [encoder/decoder](,./core/encoding.md) used by the application, used to marshal the parameters and query before making the Tendermint RPC request and unmarshal the returned response into a JSON object.
* **Account Decoder**: The account decoder from the [`auth`](.../spec/auth) module, which translates `[]byte`s into accounts.
* **RPC Client**: The [Tendermint RPC Client](https://github.com/tendermint/tendermint/blob/master/rpc/client/interface.go), or node, to which the request will be relayed to.
* **Keybase**: A [Key Manager](.//core/accounts-keys.md) used to sign transactions and handle other operations with keys.
* **Output Writer**: A [Writer](https://golang.org/pkg/io/#Writer) used to output the response.
* **Configurations**: The flags configured by the user for this command, including `--height`, specifying the height of the blockchain to query and `--indent`, which indicates to add an indent to the JSON response.

For full specification of the `CLIContext` type, click [here](https://github.com/cosmos/cosmos-sdk/blob/master/client/context/context.go#L36-L59).

### Parameters and Route Creation

The first step is to parse the command or request, extract the arguments, create a `queryRoute`, and encode everything.

**Arguments:** In this case, `Query` contains an [address](../core/accounts-fees.md) `delegatorAddress` as its only argument. However, the request can only contain `[]byte`s, as it will be relayed to a consensus engine node that has no inherent knowledge of the application types. Thus, the `CLIContext` `codec` is used to marshal the address as the type [`QueryDelegatorParams`](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/types/querier.go#L30-L38). All query arguments (e.g. the [`staking`](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/staking) module also has [`QueryValidatorParams`](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/types/querier.go#L45-L53) and [`QueryBondsParams`](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/types/querier.go#L59-L69)) have their own types that the application `codec` understands how to encode and decode. The module [`querier`](.//building-modules/querier.md) declares these types and the application registers the `codec`s.

**Route:** A `route` is also created for `Query` so that the application will understand which module to route the query to. [Baseapp](../core/baseapp.md#query-routing) will understand this query to be a `custom` query in the module `staking` with the type `QueryDelegatorDelegations`. Thus, the route will be `"custom/staking/delegatorDelegations"`.

### ABCI Query

The `CLIContext`'s main `Query` function takes the `route` and arguments. It first retrieves the RPC Client (called the [**node**](../core/node.md)) configured by the user to relay this query to, and creates the `ABCIQueryOptions` (parameters formatted for the ABCI call). The node is then used to make the ABCI call, `ABCIQueryWithOptions()`.

## Tendermint and ABCI

With a call to `ABCIQueryWithOptions()`, `Query` is received by a full-node which will then process the request. Note that, while the RPC is made to the consensus engine (e.g. Tendermint Core) of a full-node, queries are not part of consensus and will not be broadcasted to the rest of the network, as they do not require anything the network needs to agree upon.

Read more about ABCI Clients and Tendermint RPC in the Tendermint documentation [here](https://tendermint.com/rpc).

## Application Query Handling

[baseapp](../core/baseapp.md) implements the ABCI [`Query()`](../core/baseapp.md#query) function and handles four different types of queries: `app`, `store`, `p2p`, and `custom`. The `queryRoute` is parsed such that the first string must be one of the four options, then the rest of the path is parsed within the subroutines handling each type of query. The first three types (`app`, `store`, `p2p`) are purely application-level and thus directly handled by Baseapp or the stores, but the `custom` query type requires Baseapp to route the query to a module's [querier](../building-modules/querier.md).

Since `Query` is a custom query type from the `staking` module, Baseapp first parses the path, then uses the `QueryRouter` to retrieve the corresponding querier. The querier is responsible for recognizing this query, retrieving the appropriate values from the application's stores, and returning a response. Read more about queriers [here](../building-modules/querier.md).

## Response

Since `Query()` is an ABCI function, Baseapp returns the response as an [`abci.ResponseQuery`](https://tendermint.com/docs/spec/abci/abci.html#messages) type. The `CLIContext` `Query()` routine receives the response and, if `--trust-node` is toggled to `false` and a proof needs to be verified, the response is verified with the `CLIContext` [`verifyProof()`](https://github.com/cosmos/cosmos-sdk/blob/master/client/context/query.go#L136-L173) function before the response is returned.

### CLI Response

The application [`codec`](../core/encoding.md) is used to unmarshal the response to a JSON and the `CLIContext` prints the output to the command line, applying any configurations such as `--indent`.

### REST Response

The [REST server](./rest.md#rest-server) uses the `CLIContext` to format the response properly, then uses the HTTP package to write the appropriate response or error.

## Next

Read about how to build a [Command-Line Interface](./cli.md), or a [REST Interface](./rest.md).
