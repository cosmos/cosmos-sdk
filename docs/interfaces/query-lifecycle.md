<!--
order: 2
synopsis: "This document describes the lifecycle of a query in a SDK application, from the user interface to application stores and back. The query will be referred to as `Query`."
-->

# Query Lifecycle

## Pre-requisite Readings {hide}

* [Introduction to Interfaces](./interfaces-intro.md) {prereq}

## Query Creation

A [**query**](../building-modules/messages-and-queries.md#queries) is a request for information made by end-users of applications through an interface and processed by a full-node. Users can query information about the network, the application itself, and application state directly from the application's stores or modules. Note that queries are different from [transactions](../core/transactions.md) (view the lifecycle [here](../basics/tx-lifecycle.md)), particularly in that they do not require consensus to be processed (as they do not trigger state-transitions); they can be fully handled by one full-node.

For the purpose of explaining the query lifecycle, let's say `Query` is requesting a list of delegations made by a certain delegator address in the application called `app`. As to be expected, the [`staking`](https://github.com/cosmos/cosmos-sdk/tree/master/x/staking/spec) module handles this query. But first, there are a few ways `Query` can be created by users.

### CLI

The main interface for an application is the command-line interface. Users connect to a full-node and run the CLI directly from their machines - the CLI interacts directly with the full-node. To create `Query` from their terminal, users type the following command:

```bash
appcli query staking delegations <delegatorAddress>
```

This query command was defined by the [`staking`](https://github.com/cosmos/cosmos-sdk/tree/master/x/staking/spec) module developer and added to the list of subcommands by the application developer when creating the CLI. The code for this particular command is the following:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/staking/client/cli/query.go#L250-L293

Note that the general format is as follows:

```bash
appcli query [moduleName] [command] <arguments> --flag <flagArg>
```

To provide values such as `--node` (the full-node the CLI connects to), the user can use the `config` command to set themn or provide them as flags.

The CLI understands a specific set of commands, defined in a hierarchical structure by the application developer: from the [root command](./cli.md#root-command) (`appcli`), the type of command (`query`), the module that contains the command  (`staking`), and command itself (`delegations`). Thus, the CLI knows exactly which module handles this command and directly passes the call there.

### REST

Another interface through which users can make queries is through HTTP Requests to a [REST server](./rest.md#rest-server). The REST server contains, among other things, a [`CLIContext`](#clicontext) and [mux](./rest.md#gorilla-mux) router. The request looks like this:

```bash
GET http://localhost:{PORT}/staking/delegators/{delegatorAddr}/delegations
```

To provide values such as `--node` (the full-node the CLI connects to) that are required by [`baseReq`](../building-modules/module-interfaces.md#basereq), the user must configure their local REST server with the values or provide them in the request body.

The router automatically routes the `Query` HTTP request to the staking module `delegatorDelegationsHandlerFn()` function.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/staking/client/rest/query.go#L103-L106

Since this function is defined within the module and thus has no inherent knowledge of the application `Query` belongs to, it takes in the application `codec` and `CLIContext` as parameters.

To summarize, when users interact with the interfaces, they create a CLI command or HTTP request. `Query` now exists in one of these two forms, but needs to be transformed into an object understood by a full-node.

## Query Preparation

The interactions from the users' perspective are a bit different, but the underlying functions are almost identical because they are implementations of the same command defined by the module developer. This step of processing happens within the CLI or REST server and heavily involves a `CLIContext`.

### CLIContext

The first thing that is created in the execution of a CLI command is a `CLIContext`, while the REST Server directly provides a `CLIContext` for the REST Request handler. A [Context](../core/context.md) is an immutable object that stores all the data needed to process a request on the user side. In particular, a `CLIContext` stores the following:

* **Codec**: The [encoder/decoder](../core/encoding.md) used by the application, used to marshal the parameters and query before making the Tendermint RPC request and unmarshal the returned response into a JSON object.
* **Account Decoder**: The account decoder from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/spec) module, which translates `[]byte`s into accounts.
* **RPC Client**: The Tendermint RPC Client, or node, to which the request will be relayed to.
* **Keybase**: A [Key Manager](../basics/accounts.md#keybase) used to sign transactions and handle other operations with keys.
* **Output Writer**: A [Writer](https://golang.org/pkg/io/#Writer) used to output the response.
* **Configurations**: The flags configured by the user for this command, including `--height`, specifying the height of the blockchain to query and `--indent`, which indicates to add an indent to the JSON response.

The `CLIContext` also contains various functions such as `Query()` which retrieves the RPC Client and makes an ABCI call to relay a query to a full-node. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/client/context/context.go#L23-L47

The `CLIContext`'s primary role is to store data used during interactions with the end-user and provide methods to interact with this data - it is used before and after the query is processed by the full-node. Specifically, in handling `Query`, the `CLIContext` is utilized to encode the query parameters, retrieve the full-node, and write the output. Prior to being relayed to a full-node, the query needs to be encoded into a `[]byte` form, as full-nodes are application-agnostic and do not understand specific types. The full-node (RPC Client) itself is retrieved using the `CLIContext`, which knows which node the user CLI is connected to. The query is relayed to this full-node to be processed. Finally, the `CLIContext` contains a `Writer` to write output when the response is returned. These steps are further described in later sections.

### Arguments and Route Creation

At this point in the lifecycle, the user has created a CLI command or HTTP Request with all of the data they wish to include in their `Query`. A `CLIContext` exists to assist in the rest of the `Query`'s journey. Now, the next step is to parse the command or request, extract the arguments, create a `queryRoute`, and encode everything. These steps all happen on the user side within the interface they are interacting with.

#### Encoding

In this case, `Query` contains an [address](../basics/accounts.md#addresses) `delegatorAddress` as its only argument. However, the request can only contain `[]byte`s, as it will be relayed to a consensus engine (e.g. Tendermint Core) of a full-node that has no inherent knowledge of the application types. Thus, the `codec` of `CLIContext` is used to marshal the address.

Here is what the code looks like for the CLI command:

```go
delAddr, err := sdk.AccAddressFromBech32(args[0])
bz, err := cdc.MarshalJSON(types.NewQueryDelegatorParams(delAddr))
```

Here is what the code looks like for the HTTP Request:

```go
vars := mux.Vars(r)
bech32delegator := vars["delegatorAddr"]
delegatorAddr, err := sdk.AccAddressFromBech32(bech32delegator)
cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
if !ok {
	return
}
params := types.NewQueryDelegatorParams(delegatorAddr)
```

#### Query Route Creation

Important to note is that there will never be a "query" object created for `Query`; the SDK actually takes a simpler approach. Instead of an object, all the full-node needs to process a query is its `route` which specifies exactly which module to route the query to and the name of this query type. The `route` will be passed to the application `baseapp`, then module, then [querier](../building-modules/querier.md), and each will understand the `route` and pass it to the appropriate next step. [`baseapp`](../core/baseapp.md#query-routing) will understand this query to be a `custom` query in the module `staking`, and the `staking` module querier supports the type `QueryDelegatorDelegations`. Thus, the route will be `"custom/staking/delegatorDelegations"`.

Here is what the code looks like:

```go
route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorDelegations)
```

Now, `Query` exists as a set of encoded arguments and a route to a specific module and its query type. It is ready to be relayed to a full-node.

#### ABCI Query Call

The `CLIContext` has a `Query()` function used to retrieve the pre-configured node and relay a query to it; the function takes the query `route` and arguments as parameters. It first retrieves the RPC Client (called the [**node**](../core/node.md)) configured by the user to relay this query to, and creates the `ABCIQueryOptions` (parameters formatted for the ABCI call). The node is then used to make the ABCI call, `ABCIQueryWithOptions()`.

Here is what the code looks like:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/client/context/query.go#L75-L112

## RPC

With a call to `ABCIQueryWithOptions()`, `Query` is received by a [full-node](../core/encoding.md) which will then process the request. Note that, while the RPC is made to the consensus engine (e.g. Tendermint Core) of a full-node, queries are not part of consensus and will not be broadcasted to the rest of the network, as they do not require anything the network needs to agree upon.

Read more about ABCI Clients and Tendermint RPC in the Tendermint documentation [here](https://tendermint.com/rpc).

## Application Query Handling

When a query is received by the full-node after it has been relayed from the underlying consensus engine, it is now being handled within an environment that understands application-specific types and has a copy of the state. [`baseapp`](../core/baseapp.md) implements the ABCI [`Query()`](../core/baseapp.md#query) function and handles four different types of queries: `app`, `store`, `p2p`, and `custom`. The `queryRoute` is parsed such that the first string must be one of the four options, then the rest of the path is parsed within the subroutines handling each type of query. The first three types (`app`, `store`, `p2p`) are purely application-level and thus directly handled by `baseapp` or the stores, but the `custom` query type requires `baseapp` to route the query to a module's [querier](../building-modules/querier.md).

Since `Query` is a custom query type from the `staking` module, `baseapp` first parses the path, then uses the `QueryRouter` to retrieve the corresponding querier, and routes the query to the module. The querier is responsible for recognizing this query, retrieving the appropriate values from the application's stores, and returning a response. Read more about queriers [here](../building-modules/querier.md).

Once a result is received from the querier, `baseapp` begins the process of returning a response to the user.

## Response

Since `Query()` is an ABCI function, `baseapp` returns the response as an [`abci.ResponseQuery`](https://tendermint.com/docs/spec/abci/abci.html#messages) type. The `CLIContext` `Query()` routine receives the response and, if `--trust-node` is toggled to `false` and a proof needs to be verified, the response is verified with the `CLIContext` `verifyProof()` function before the response is returned.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/client/context/query.go#L127-L165

### CLI Response

The application [`codec`](../core/encoding.md) is used to unmarshal the response to a JSON and the `CLIContext` prints the output to the command line, applying any configurations such as `--indent`. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/staking/client/cli/query.go#L252-L293

### REST Response

The [REST server](./rest.md#rest-server) uses the `CLIContext` to format the response properly, then uses the HTTP package to write the appropriate response or error. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/staking/client/rest/utils.go#L115-L148

## Next {hide}

Read about how to build a [Command-Line Interface](./cli.md), or a [REST Interface](./rest.md) {hide}
