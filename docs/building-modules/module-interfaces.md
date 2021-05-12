<!--
order: 11
-->

# Module Interfaces

This document details how to build CLI and REST interfaces for a module. Examples from various SDK modules are included. {synopsis}

## Pre-Requisite Readings

- [Building Modules Intro](./intro.md) {prereq}

## CLI

One of the main interfaces for an application is the [command-line interface](../interfaces/cli.md). This entrypoint adds commands from the application's modules enabling end-users to create [**messages**](./messages-and-queries.md#messages) (wrapped in transactions) and [**queries**](./messages-and-queries.md#queries). The CLI files are typically found in the module's `./client/cli` folder.

### Transaction Commands

 In order to create messages that trigger state changes, end-users must create [transactions](../core/transactions.md) that wrap and deliver the messages. A transaction command creates a transaction that includes one or more messages.
 
 Transaction commands typically have their own `tx.go` file that lives within the module's `./client/cli` folder. The commands are specified in getter functions and typically the name of the function includes the name of the command.

Here is an example from the `x/bank` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/x/bank/client/cli/tx.go#L28-L63

In the example above, `NewSendTxCmd` creates and returns the transaction command for a transaction that wraps and delivers `MsgSend`. `MsgSend` is the message used to send tokens from one account to another.

In general, the getter function does the following:

- **Constructs the command:** Read the [Cobra Documentation](https://godoc.org/github.com/spf13/cobra) for more detailed information on how to create commands.
  - **Use:** Specifies the format of the user input required to invoke the command. In the example above, `send` is the name of the transaction command and `[from_key_or_address]`, `[to_address]`, and `[amount]` are the arguments.
  - **Args:** The number of arguments the user provides. In this case, there are exactly three: `[from_key_or_address]`, `[to_address]`, and `[amount]`.
  - **Short and Long:** Descriptions for the command. A `Short` description is expected. A `Long` description is available for additional information that is provided when a user adds the `--help` flag.
  - **RunE:** Defines a function that can return an error. This is the function that is called when the command is executed. This function encapsulates all of the logic to create a new transaction.
    - In general, the function usually starts by getting the `clientCtx` with `client.GetClientTxContext(cmd)`. The `clientCtx` contains information and helper methods relevant to transaction handling, including information about the user. In this example, the `clientCtx` is used to retrieve the address of the sender by calling `clientCtx.GetFromAddress()`.
    - If applicable, the command's arguments are parsed. In this example, the arguments `[to_address]` and `[amount]` are both parsed.
    - A [message](./messages-and-queries.md) is created using the parsed arguments and information from the `clientCtx`. The constructor function of the message type is called directly. In this case, `types.NewMsgSend(fromAddr, toAddr, amount)`. Its good practice to call `msg.ValidateBasic()` after creating the message, which runs a sanity check on the provided arguments.
    - Depending on what the user wants, the transaction is either generated offline or signed and broadcasted to the preconfigured node using `GenerateOrBroadcastTxCLI(clientCtx, flags, msg)`.
- **Adds transaction flags:** All transaction commands must add a set of transaction [flags](#flags). The transaction flags are used to collect additional information from the user (e.g. the amount of fees the user is willing to pay). The transaction flags are added to the constructed command using `AddTxFlagsToCmd(cmd)`.
- **Adds additional flags:** Some transaction commands may require additional flags that are specific to the command. See [flags](#flags) for more information.
- **Returns the command:** Finally, the transaction command is returned.

Each module needs to have a `GetTxCmd()`, which aggregates all of the transaction commands of the module. Application developers wishing to include the module's transactions will call this function to add them as subcommands in the CLI. Here is the `GetTxCmd()` function for the `x/auth` module, which adds the `Sign`, `MultiSign`, `ValidateSignatures` and `SignBatch` transaction commands.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/x/auth/client/cli/tx.go#L10-L26

An application using the `x/auth` module can then add the aggregated transaction commands to the root command by calling `rootCmd.AddCommand(auth.GetTxCmd())`.

### Query Commands

[Queries](./messages-and-queries.md#queries) allow users to gather information about the application or network state; they are routed by the application and processed by the module in which they are defined. Query commands typically have their own `query.go` file in the module's `./client/cli` folder. Like transaction commands, they are specified in getter functions. Here is an example of a query command from the `x/auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.4/x/auth/client/cli/query.go#L75-L105

This query returns the account at a given address. The getter function does the following:

- **Constructs the command:** Read the [Cobra Documentation](https://godoc.org/github.com/spf13/cobra) for more detailed information on how to create commands.
  - **Use:** Specifies the format of the user input required to invoke the command. In the example above, `account` is the name of the query command and `[address]` is the argument.
  - **Args:** The number of arguments the user provides. In this case, there is exactly one: `[address]`.
  - **Short and Long:** Descriptions for the command. A `Short` description is expected. A `Long` description is available for additional information that is provided when a user adds the `--help` flag.
  - **RunE:** Defines a function that can return an error. This is the function that is called when the command is executed. This function encapsulates all of the logic to create a new query.
    - In general, the function usually starts by getting the `clientCtx` with `client.GetClientQueryContext(cmd)`. The `clientCtx` contains information and helper methods relevant to query handling.
    - If applicable, the command's arguments are parsed. In this example, the argument `[address]` is parsed.
    - A new `queryClient` is initialized using `NewQueryClient(clientCtx)`. The `queryClient` is then used to call the appropriate [query](./messages-and-queries.md#grpc-queries).
    - The `clientCtx.PrintProto` method is used to format the `proto.Message` object so that the results can be printed back to the user.
- **Adds query flags:** All query commands must add a set of query [flags](#flags). The query flags are added to the constructed command using `AddQueryFlagsToCmd(cmd)`.
- **Adds additional flags:** Some query commands may require additional flags that are specific to the command. See [flags](#flags) for more information.
- **Returns the command:** Finally, the query command is returned.

Each module needs to have a `GetQueryCmd()`, which aggregates all of the query commands of the module. Application developers wishing to include the module's queries will call this function to add them as subcommands in their CLI. Its structure is identical to the `GetTxCmd()` command shown above.

### Flags

[Flags](../interfaces/cli.md#flags) are entered by the user and allow for command customizations. Examples include the [fees](../basics/gas-fees.md) or gas prices users are willing to pay for their transactions.

The flags for a module are typically found in a `flags.go` file in the module's `./client/cli` folder. Module developers can create a list of possible flags including the value type, default value, and a description displayed if the user uses a `help` command. In each transaction getter function, they can add flags to the commands and, optionally, mark flags as _required_ so that an error is thrown if the user does not provide values for them.

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

In order to do that, modules should implement `RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux)` on `AppModuleBasic` to wire the client gRPC requests to the correct handler inside the module.

Here's an example from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/module.go#L69-L72

## gRPC-gateway REST

Applications typically support web services that use HTTP requests (e.g. a web wallet like [Lunie.io](https://lunie.io)). Thus, application developers can also use REST Routes to route HTTP requests to the application's modules; these routes will be used by service providers.

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

With this implementation, module developers need to define the REST client by defining [routes](#register-routes) for all possible [requests](#request-types) and [handlers](#request-handlers) for each of them. It's up to the module developer how to organize the REST interface files; there is typically a `rest.go` file found in the module's `./client/rest` folder.

To support HTTP requests, the module developer needs to define possible request types, how to handle them, and provide a way to register them with a provided router.

### Request Types

Request types, which define structured interactions from users, must be defined for all _transaction_ requests. Users using this method to interact with an application will send HTTP Requests with the required fields in order to trigger state changes in the application. Conventionally, each request is named with the suffix `Req`, e.g. `SendReq` for a Send transaction. Each struct should include a base request [`baseReq`](../interfaces/rest.md#basereq), the name of the transaction, and all the arguments the user must provide for the transaction.

Here is an example of a request to send coins from the `x/bank` module:

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
- `Gas` refers to how much [gas](../basics/gas-fees.md), which represents computational resources, a transaction consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing `auto` as the value for `Gas`.
- `GasAdjustment` can be used to scale gas up in order to avoid underestimating. For example, users can specify their gas adjustment as `1.5` to use 1.5 times the estimated gas.
- `GasPrices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, `--gas-prices=0.025uatom, 0.025upho` means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
- `Fees` specifies how much in [fees](../basics/gas-fees.md) the user is willing to pay in total. Note that the user only needs to provide either `gas-prices` or `fees`, but not both, because they can be derived from each other.
- `Simulate` instructs the application to ignore gas and simulate the transaction running without broadcasting.

### Request Handlers

Request handlers must be defined for both transaction and query requests. Handlers' arguments include a reference to the [client `Context`](../interfaces/query-lifecycle.md#context).

Here is an example of a request handler for the `x/bank` module `SendReq` request (the same one shown above):

+++ https://github.com/cosmos/cosmos-sdk/blob/7f59723d889b69ca19966167f0b3a7fec7a39e53/x/bank/client/rest/tx.go#L21-L51

The request handler can be broken down as follows:

- **Parse Request:** First, the request handler tries to parse the argument `address` into a `AccountAddress`. The request handler then attempts to parse the request, followed by running `Sanitize` and `ValidateBasic` on the underlying `BaseReq` to check the validity of the request. Finally, the request handler attempts to parse `BaseReq.From` to the type `AccountAddress`.
- **Message:** After parsing the request, a [message](./messages-and-queries.md#messages) of type `MsgSend` is created from the values, which is defined by the module developer to trigger the state changes for this transaction.
- **Generate Transaction:** Finally, the client `Context`, the HTTP `ResponseWriter`, the request's [`BaseReq`](../interfaces/rest.md#basereq), and the message are all passed to `WriteGeneratedTxResponse` to further process the request.

To read more about how a transaction is generated, visit the transactions documentation [here](../core/transactions.md#transaction-generation).

### Register Routes

The application CLI entrypoint will have a `RegisterRoutes` function in its `main.go` file, which calls the `registerRoutes` functions of each module utilized by the application. Module developers need to implement `registerRoutes` for their modules so that applications are able to route messages and queries to their corresponding handlers and queriers.

The router used by the SDK is [Gorilla Mux](https://github.com/gorilla/mux). The router is initialized with the Gorilla Mux `NewRouter()` function. The router's `HandleFunc` function can then be used to route urls with the defined request handlers and the HTTP method (e.g. "POST", "GET") as a route matcher. It is recommended to prefix every route with the name of the module to avoid collisions with other modules that have the same query or transaction names.

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
