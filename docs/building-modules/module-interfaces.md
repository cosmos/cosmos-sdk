<!--
order: 11
-->

# Module Interfaces

This document details how to build CLI and REST interfaces for a module. Examples from various SDK modules are included. {synopsis}

## Pre-requisite Readings

- [Building Modules Intro](./intro.md) {prereq}

## CLI

One of the main interfaces for an application is the [command-line interface](../core/cli.md). This entrypoint adds commands from the application's modules to let end-users create [**messages**](./messages-and-queries.md#messages) and [**queries**](./messages-and-queries.md#queries). The CLI files are typically found in the `./x/moduleName/client/cli` folder.

### Transaction Commands

[Transactions](../core/transactions.md) are created by users to wrap messages that trigger state changes when they get included in a valid block. Transaction commands typically have their own `tx.go` file in the module `./x/moduleName/client/cli` folder. The commands are specified in getter functions and include the name of the command.

Here is an example from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/client/cli/tx_sign.go#L160-L194

This getter function creates the command for the `Sign` transaction. It does the following:

- **Construct the command:** Read the [Cobra Documentation](https://godoc.org/github.com/spf13/cobra) for details on how to create commands.
  - **Use:** Specifies the format of a command-line entry users should type in order to invoke this command. In this case, the user uses `buy-name` as the name of the transaction command and provides the `name` the user wishes to buy and the `amount` the user is willing to pay.
  - **Args:** The number of arguments the user provides, in this case exactly two: `name` and `amount`.
  - **Short and Long:** A description for the function is provided here. A `Short` description is expected, and `Long` can be used to provide a more detailed description when a user uses the `--help` flag to ask for more information.
  - **RunE:** Defines a function that can return an error, called when the command is executed. Using `Run` would do the same thing, but would not allow for errors to be returned.
- **`RunE` Function Body:** The function should be specified as a `RunE` to allow for errors to be returned. This function encapsulates all of the logic to create a new transaction that is ready to be relayed to nodes.
  - The function should first get the `clientCtx` with `client.GetClientContextFromCmd(cmd)` and `client.ReadTxCommandFlags(clientCtx, cmd.Flags())`. This context contains all the information provided by the user and will be used to transfer this user-specific information between processes. To learn more about how contexts are used in a transaction, click [here](../core/transactions.md#transaction-generation).
  - If applicable, the command's arguments are parsed.
  - If applicable, the `Context` is used to retrieve any parameters such as the transaction originator's address to be used in the transaction. Here, the `from` address is retrieved by calling `clientCtx.GetFromAddress()`.
  - A [message](./messages-and-queries.md) is created using all parameters parsed from the command arguments and `Context`. The constructor function of the specific message type is called directly. It is good practice to call `ValidateBasic()` on the newly created message to run a sanity check and check for invalid arguments.
  - Depending on what the user wants, the transaction is either generated offline or signed and broadcasted to the preconfigured node using `GenerateOrBroadcastMsgs()`.
- **Flags.** Add any [flags](#flags) to the command. All transaction commands have flags to provide additional information from the user (e.g. amount of fees they are willing to pay). These _persistent_ [transaction flags](../interfaces/cli.md#flags) can be added to a higher-level command so that they apply to all transaction commands.

Finally, the module needs to have a `GetTxCmd()`, which aggregates all of the transaction commands of the module. Often, each command getter function has its own file in the module's `cli` folder, and a separate `tx.go` file contains `GetTxCmd()`. Application developers wishing to include the module's transactions will call this function to add them as subcommands in their CLI. Here is the `auth` `GetTxCmd()` function, which adds the `Sign`, `MultiSign`, `ValidateSignatures` and `SignBatch` commands.

+++ https://github.com/cosmos/cosmos-sdk/blob/351192aa0b52a42b66ff06e81cfa7a9e26667a7f/x/auth/client/cli/tx.go#L10-L26

An application using this module likely adds `auth` module commands to its root `TxCmd` command by calling `txCmd.AddCommand(authModuleClient.GetTxCmd())`.

### Query Commands

[Queries](./messages-and-queries.md#queries) allow users to gather information about the application or network state; they are routed by the application and processed by the module in which they are defined. Query commands typically have their own `query.go` file in the module `x/moduleName/client/cli` folder. Like transaction commands, they are specified in getter functions. Here is an example of a query command from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/d55c1a26657a0af937fa2273b38dcfa1bb3cff9f/x/auth/client/cli/query.go#L76-L108

This query returns the account at a given address. The getter function does the following:

- **Construct the command.** Read the [Cobra Documentation](https://godoc.org/github.com/spf13/cobra) and the [transaction command](#transaction-commands) example above for more information. The user must type `account` and provide the `address` they are querying for as the only argument.
- **`RunE`.** The function should be specified as a `RunE` to allow for errors to be returned. This function encapsulates all of the logic to create a new query that is ready to be relayed to nodes.
  - The function should first initialize a new client [`Context`](../interfaces/query-lifecycle.md#context) as described in the [previous section](#transaction-commands)
  - If applicable, the `Context` is used to retrieve any parameters (e.g. the query originator's address to be used in the query) and marshal them with the query parameter type, in preparation to be relayed to a node. There are no `Context` parameters in this case because the query does not involve any information about the user.
  - A new `queryClient` should be initialized using `NewQueryClient(clientCtx)`, this method being generated from `query.proto`. Then it can be used to call the appropriate [query](./messages-and-queries.md#grpc-queries).
  - The `clientCtx.PrintProto` method is used to format a `proto.Message` object and print it back to the user.
- **Flags.** Add any [flags](#flags) to the command.

Finally, the module also needs a `GetQueryCmd`, which aggregates all of the query commands of the module. Application developers wishing to include the module's queries will call this function to add them as subcommands in their CLI. Its structure is identical to the `GetTxCmd` command shown above.

### Flags

[Flags](../interfaces/cli.md#flags) are entered by the user and allow for command customizations. Examples include the [fees](../basics/gas-fees.md) or gas prices users are willing to pay for their transactions.

The flags for a module are typically found in a `flags.go` file in the `./x/moduleName/client/cli` folder. Module developers can create a list of possible flags including the value type, default value, and a description displayed if the user uses a `help` command. In each transaction getter function, they can add flags to the commands and, optionally, mark flags as _required_ so that an error is thrown if the user does not provide values for them.

For full details on flags, visit the [Cobra Documentation](https://github.com/spf13/cobra).

For example, the SDK `./client/flags` package includes a `AddTxFlagsToCmd(cmd *cobra.Command)` function that adds necessary flags to a transaction command, such as the `from` flag to indicate which address the transaction originates from.

+++ https://github.com/cosmos/cosmos-sdk/blob/cfb5fc03e5092395403d10156c0ee96e6ff1ddbe/client/flags/flags.go#L85-L112

Here is an example of how to add a flag using the `from` flag from this function.

```go
cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
```

The input provided for this flag - called `FlagFrom` is a string with the default value of `""` if none is provided. If the user asks for a description of this flag, the description will be printed.

A flag can be marked as _required_ so that an error is automatically thrown if the user does not provide a value:

```go
cmd.MarkFlagRequired(FlagFrom)
```

Since `AddTxFlagsToCmd(cmd *cobra.Command)` includes all of the basic flags required for a transaction command, module developers may choose not to add any of their own (specifying arguments instead may often be more appropriate).

Similarly, there is a `AddQueryFlagsToCmd(cmd *cobra.Command)` to add common flags to a module query command.

## gRPC

[gRPC](https://grpc.io/) is the prefered way for external clients like wallets and exchanges to interact with a node.

In addition to providing an ABCI query pathway, modules [custom queries](./messages-and-queries.md#grpc-queries) can provide a GRPC proxy server that routes requests in the GRPC protocol to ABCI query requests under the hood.

In order to do that, module should implement `RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux)` on `AppModuleBasic` to wire the client gRPC requests to the correct handler inside the module.

Here's an example from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/module.go#L69-L72

## gRPC-gateway REST

Applications typically support web services that use HTTP requests (e.g. a web wallet like [Lunie.io](https://lunie.io). Thus, application developers can also use REST Routes to route HTTP requests to the application's modules; these routes will be used by service providers.

[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) translates REST calls into gRPC calls, which might be useful for clients that do not use gRPC.

Modules that want to expose REST queries should add `google.api.http` annotations to their `rpc` methods, such as in the example below from the `auth` module:

```proto
// Query defines the gRPC querier service.
service Query{
    // Account returns account details based on address.
    rpc Account (QueryAccountRequest) returns (QueryAccountResponse) {
      option (google.api.http).get = "/cosmos/auth/v1beta1/accounts/{address}";
    }

    // Params queries all parameters.
    rpc Params (QueryParamsRequest) returns (QueryParamsResponse) {
      option (google.api.http).get = "/cosmos/auth/v1beta1/params";
    }
}
```

gRPC gateway is started in-process along with the application and Tendermint. It can be enabled or disabled by setting gRPC Configuration `enable` in [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml).

The SDK provides a command for generating [Swagger](https://swagger.io/) documentation (`protoc-gen-swagger`). Setting `swagger` in [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml) defines if swagger documentation should be automatically registered.

## Legacy REST

Legacy REST endpoints will be deprecated. But developers may choose to keep using legacy REST endpoints for backward compatibility, although the recommended way is to use [gRPC](#grpc) and [gRPC-gateway](#grpc-gateway-rest).

With this implementation, module developers need to define the REST client by defining [routes](#register-routes) for all possible [requests](#request-types) and [handlers](#request-handlers) for each of them. It's up to the module developer how to organize the REST interface files; there is typically a `rest.go` file found in the module's `./x/moduleName/client/rest` folder.

To support HTTP requests, the module developer needs to define possible request types, how to handle them, and provide a way to register them with a provided router.

### Request Types

Request types, which define structured interactions from users, must be defined for all _transaction_ requests. Users using this method to interact with an application will send HTTP Requests with the required fields in order to trigger state changes in the application. Conventionally, each request is named with the suffix `Req`, e.g. `SendReq` for a Send transaction. Each struct should include a base request [`baseReq`](../interfaces/rest.md#basereq), the name of the transaction, and all the arguments the user must provide for the transaction.

Here is an example of a request to send coins from the `bank` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/7f59723d889b69ca19966167f0b3a7fec7a39e53/x/bank/client/rest/tx.go#L15-L19

The `BaseReq` includes basic information that every request needs to have, similar to required flags in a CLI. All of these values, including `GasPrices` and `AccountNumber`, will be provided in the request body. The user will also need to specify the argument `Amount` fields in the body.

#### BaseReq

`BaseReq` is a type defined in the SDK that encapsulates much of the transaction configurations similar to CLI command flags. Users must provide the information in the body of their requests.

- `From` indicates which [account](../basics/accounts.md) the transaction originates from. This account is used to sign the transaction.
- `Memo` sends a memo along with the transaction.
- `ChainID` specifies the unique identifier of the blockchain the transaction pertains to.
- `AccountNumber` is an identifier for the account.
- `Sequence`is the value of a counter measuring how many transactions have been sent from the account. It is used to prevent replay attacks.
- `TimeoutHeight` allows a transaction to be rejected if it's committed at a height greater than the timeout.
- `Gas` refers to how much [gas](../basics/gas-fees.md), which represents computational resources, Tx consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing auto as the value for `Gas`.
- `GasAdjustment` can be used to scale gas up in order to avoid underestimating. For example, users can specify their gas adjustment as 1.5 to use 1.5 times the estimated gas.
- `GasPrices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, --gas-prices=0.025uatom, 0.025upho means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
- `Fees` specifies how much in [fees](../basics/gas-fees.md) the user is willing to pay in total. Note that the user only needs to provide either `gas-prices` or `fees`, but not both, because they can be derived from each other.
- `Simulate` instructs the application to ignore gas and simulate the transaction running without broadcasting.

### Request Handlers

Request handlers must be defined for both transaction and query requests. Handlers' arguments include a reference to the [client `Context`](../interfaces/query-lifecycle.md#context).

Here is an example of a request handler for the `bank` module `SendReq` request (the same one shown above):

+++ https://github.com/cosmos/cosmos-sdk/blob/7f59723d889b69ca19966167f0b3a7fec7a39e53/x/bank/client/rest/tx.go#L21-L51

The request handler can be broken down as follows:

- **Parse Request:** First, it tries to parse the argument `address` into a `AccountAddress`. Then, the request handler attempts to parse the request, and then run `Sanitize` and `ValidateBasic` on the underlying `BaseReq` to check the validity of the request. Finally, it attempts to parse `BaseReq.From` to the type `AccountAddress`.
- **Message:** Then, a [message](./messages-and-queries.md#messages) of the type `MsgSend` (defined by the module developer to trigger the state changes for this transaction) is created from the values.
- **Generate Transaction:** Finally, the HTTP `ResponseWriter`, client `Context`, request [`BaseReq`](../interfaces/rest.md#basereq), and message is passed to `WriteGeneratedTxResponse` to further process the request.

To read more about how a transaction is generated, visit the transactions documentation [here](../core/transactions.md#transaction-generation).

### Register Routes

The application CLI entrypoint will have a `RegisterRoutes` function in its `main.go` file, which calls the `registerRoutes` functions of each module utilized by the application. Module developers need to implement `registerRoutes` for their modules so that applications are able to route messages and queries to their corresponding handlers and queriers.

The router used by the SDK is [Gorilla Mux](https://github.com/gorilla/mux). The router is initialized with the Gorilla Mux `NewRouter()` function. Then, the router's `HandleFunc` function can then be used to route urls with the defined request handlers and the HTTP method (e.g. "POST", "GET") as a route matcher. It is recommended to prefix every route with the name of the module to avoid collisions with other modules that have the same query or transaction names.

Here is a `registerRoutes` function with one query route example from the [nameservice tutorial](https://cosmos.network/docs/tutorial/rest.html):

```go
func RegisterRoutes(cliCtx client.Context, r *mux.Router, cdc *codec.LegacyAmino, storeName string) {
  // ResolveName Query
  r.HandleFunc(fmt.Sprintf("/%s/names/{%s}", storeName, restName), resolveNameHandler(cdc, cliCtx, storeName)).Methods("GET")
}
```

A few things to note:

- The router `r` has already been initialized by the application and is passed in here as an argument - this function is able to add on the nameservice module's routes onto any application's router. The application must also provide a [`Context`](../interfaces/query-lifecycle.md#context) that the querier will need to process user requests and the application [`codec`](../core/encoding.md) for encoding and decoding application-specific types.
- `"/%s/names/{%s}", storeName, restName` is the url for the HTTP request. `storeName` is the name of the module, `restName` is a variable provided by the user to specify what kind of query they are making.
- `resolveNameHandler` is the query request handler defined by the module developer. It also takes the application `codec` and `Context` passed in from the user side, as well as the `storeName`.
- `"GET"` is the HTTP Request method. As to be expected, queries are typically GET requests. Transactions are typically POST and PUT requests.

## Next {hide}

Read about the recommended [module structure](./structure.md) {hide}
