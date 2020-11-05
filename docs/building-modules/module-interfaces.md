<!--
order: 11
synopsis: This document details how to build CLI and REST interfaces for a module. Examples from various SDK modules are included.
-->

# Module Interfaces

## Pre-requisite Readings {hide}

* [Building Modules Intro](./intro.md) {prereq}

## CLI

One of the main interfaces for an application is the [command-line interface](../interfaces/cli.md). This entrypoint adds commands from the application's modules to let end-users create [**messages**](./messages-and-queries.md#messages) and [**queries**](./messages-and-queries.md#queries).  The CLI files are typically found in the `./x/moduleName/client/cli` folder.

### Transaction Commands

[Transactions](../core/transactions.md) are created by users to wrap messages that trigger state changes when they get included in a valid block. Transaction commands typically have their own `tx.go` file in the module `./x/moduleName/client/cli` folder. The commands are specified in getter functions prefixed with `GetCmd` and include the name of the command. 

Here is an example from the `nameservice` module:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/client/cli/tx.go#L33-L58

This getter function creates the command for the Buy Name transaction. It does the following:

- **Construct the command:** Read the [Cobra Documentation](https://github.com/spf13/cobra) for details on how to create commands.
  + **Use:** Specifies the format of a command-line entry users should type in order to invoke this command. In this case, the user uses `buy-name` as the name of the transaction command and provides the `name` the user wishes to buy and the `amount` the user is willing to pay.
  + **Args:** The number of arguments the user provides, in this case exactly two: `name` and `amount`.
  + **Short and Long:** A description for the function is provided here. A `Short` description is expected, and `Long` can be used to provide a more detailed description when a user uses the `--help` flag to ask for more information.
  + **RunE:** Defines a function that can return an error, called when the command is executed. Using `Run` would do the same thing, but would not allow for errors to be returned.
- **`RunE` Function Body:** The function should be specified as a `RunE` to allow for errors to be returned. This function encapsulates all of the logic to create a new transaction that is ready to be relayed to nodes.
  + The function should first initialize a [`TxBuilder`](../core/transactions.md#txbuilder) with the application `codec`'s `TxEncoder`, as well as a new [`CLIContext`](../interfaces/query-lifecycle.md#clicontext) with the `codec` and `AccountDecoder`. These contexts contain all the information provided by the user and will be used to transfer this user-specific information between processes. To learn more about how contexts are used in a transaction, click [here](../core/transactions.md#transaction-generation).
  + If applicable, the command's arguments are parsed. Here, the `amount` given by the user is parsed into a denomination of `coins`.
  + If applicable, the `CLIContext` is used to retrieve any parameters such as the transaction originator's address to be used in the transaction. Here, the `from` address is retrieved by calling `cliCtx.getFromAddress()`.
  + A [message](./messages-and-queries.md) is created using all parameters parsed from the command arguments and `CLIContext`. The constructor function of the specific message type is called directly. It is good practice to call `ValidateBasic()` on the newly created message to run a sanity check and check for invalid arguments.
  + Depending on what the user wants, the transaction is either generated offline or signed and broadcasted to the preconfigured node using `GenerateOrBroadcastMsgs()`.
- **Flags.** Add any [flags](#flags) to the command. No flags were specified here, but all transaction commands have flags to provide additional information from the user (e.g. amount of fees they are willing to pay). These *persistent* [transaction flags](../interfaces/cli.md#flags) can be added to a higher-level command so that they apply to all transaction commands.

Finally, the module needs to have a `GetTxCmd()`, which aggregates all of the transaction commands of the module. Often, each command getter function has its own file in the module's `cli` folder, and a separate `tx.go` file contains `GetTxCmd()`. Application developers wishing to include the module's transactions will call this function to add them as subcommands in their CLI. Here is the `auth` `GetTxCmd()` function, which adds the `Sign` and `MultiSign` commands. 

+++ https://github.com/cosmos/cosmos-sdk/blob/67f6b021180c7ef0bcf25b6597a629aca27766b8/x/auth/client/cli/tx.go#L11-L25

An application using this module likely adds `auth` module commands to its root `TxCmd` command by calling `txCmd.AddCommand(authModuleClient.GetTxCmd())`.

### Query Commands

[Queries](./messages-and-queries.md#queries) allow users to gather information about the application or network state; they are routed by the application and processed by the module in which they are defined. Query commands typically have their own `query.go` file in the module `x/moduleName/client/cli` folder. Like transaction commands, they are specified in getter functions and have the prefix `GetCmdQuery`. Here is an example of a query command from the `nameservice` module:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/client/cli/query.go#L52-L73

This query returns the address that owns a particular name. The getter function does the following:

- **`codec` and `queryRoute`.** In addition to taking in the application `codec`, query command getters also take a `queryRoute` used to construct a path [Baseapp](../core/baseapp.md#query-routing) uses to route the query in the application.
- **Construct the command.** Read the [Cobra Documentation](https://github.com/spf13/cobra) and the [transaction command](#transaction-commands) example above for more information. The user must type `whois` and provide the `name` they are querying for as the only argument.
- **`RunE`.** The function should be specified as a `RunE` to allow for errors to be returned. This function encapsulates all of the logic to create a new query that is ready to be relayed to nodes.
  + The function should first initialize a new [`CLIContext`](../interfaces/query-lifecycle.md#clicontext) with the application `codec`.
  + If applicable, the `CLIContext` is used to retrieve any parameters (e.g. the query originator's address to be used in the query) and marshal them with the query parameter type, in preparation to be relayed to a node. There are no `CLIContext` parameters in this case because the query does not involve any information about the user.
  + The `queryRoute` is used to construct a `path` [`baseapp`](../core/baseapp.md) will use to route the query to the appropriate [querier](./querier.md). The expected format for a query `path` is "queryCategory/queryRoute/queryType/arg1/arg2/...", where `queryCategory` can be `p2p`, `store`, `app`, or `custom`, `queryRoute` is the name of the module, and `queryType` is the name of the query type defined within the module. [`baseapp`](../core/baseapp.md) can handle each type of `queryCategory` by routing it to a module querier or retrieving results directly from stores and functions for querying peer nodes. Module queries are `custom` type queries (some SDK modules have exceptions, such as `auth` and `gov` module queries).
  + The `CLIContext` `QueryWithData()` function is used to relay the query to a node and retrieve the response. It requires the `path`. It returns the result and height of the query upon success or an error if the query fails. In addition, it will verify the returned proof if `TrustNode` is disabled. If proof verification fails or the query height is invalid, an error will be returned.
  + The `codec` is used to nmarshal the response and the `CLIContext` is used to print the output back to the user.
- **Flags.** Add any [flags](#flags) to the command.


Finally, the module also needs a `GetQueryCmd`, which aggregates all of the query commands of the module. Application developers wishing to include the module's queries will call this function to add them as subcommands in their CLI. Its structure is identical to the `GetTxCmd` command shown above.

### Flags

[Flags](../interfaces/cli.md#flags) are entered by the user and allow for command customizations. Examples include the [fees](../basics/gas-fees.md) or gas prices users are willing to pay for their transactions.

The flags for a module are typically found in a `flags.go` file in the `./x/moduleName/client/cli` folder. Module developers can create a list of possible flags including the value type, default value, and a description displayed if the user uses a `help` command. In each transaction getter function, they can add flags to the commands and, optionally, mark flags as *required* so that an error is thrown if the user does not provide values for them.

For full details on flags, visit the [Cobra Documentation](https://github.com/spf13/cobra).

For example, the SDK `./client/flags` package includes a `PostCommands()` function that adds necessary flags to transaction commands, such as the `from` flag to indicate which address the transaction originates from. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/client/flags/flags.go#L85-L116

Here is an example of how to add a flag using the `from` flag from this function.

```go
cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
```

The input provided for this flag - called `FlagFrom` is a string with the default value of `""` if none is provided. If the user asks for a description of this flag, the description will be printed.

A flag can be marked as *required* so that an error is automatically thrown if the user does not provide a value:

```go
cmd.MarkFlagRequired(FlagFrom)
```

Since `PostCommands()` includes all of the basic flags required for a transaction command, module developers may choose not to add any of their own (specifying arguments instead may often be more appropriate). For a full list of what flags are included in the `PostCommands()` function, including which are required inputs from users, see the CLI documentation [here](../interfaces/cli.md#transaction-flags).

## REST

Applications typically support web services that use HTTP requests (e.g. a web wallet like [Lunie.io](https://lunie.io). Thus, application developers will also use REST Routes to route HTTP requests to the application's modules; these routes will be used by service providers. The module developer's responsibility is to define the REST client by defining [routes](#register-routes) for all possible [requests](#request-types) and [handlers](#request-handlers) for each of them. It's up to the module developer how to organize the REST interface files; there is typically a `rest.go` file found in the module's `./x/moduleName/client/rest` folder.

To support HTTP requests, the module developer needs to define possible request types, how to handle them, and provide a way to register them with a provided router.

### Request Types

Request types, which define structured interactions from users, must be defined for all *transaction* requests. Users using this method to interact with an application will send HTTP Requests with the required fields in order to trigger state changes in the application. Conventionally, each request is named with the suffix `Req`, e.g. `SendReq` for a Send transaction. Each struct should include a base request [`baseReq`](../interfaces/rest.md#basereq), the name of the transaction, and all the arguments the user must provide for the transaction.

Here is an example of a request to buy a name from the `nameservice` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/7f59723d889b69ca19966167f0b3a7fec7a39e53/x/bank/client/rest/tx.go#L15-L19

The `BaseReq` includes basic information that every request needs to have, similar to required flags in a CLI. All of these values, including `GasPrices` and `AccountNumber`, will be provided in the request body. The user will also need to specify the arguments `Name` and `Amount` fields in the body and `Buyer` will be provided by the user's address.

#### BaseReq

`BaseReq` is a type defined in the SDK that encapsulates much of the transaction configurations similar to CLI command flags. Users must provide the information in the body of their requests.

* `From` indicates which [account](../basics/accounts.md) the transaction originates from. This account is used to sign the transaction.
*	`Memo` sends a memo along with the transaction.
*	`ChainID` specifies the unique identifier of the blockchain the transaction pertains to.
*	`AccountNumber` is an identifier for the account.
*	`Sequence`is the value of a counter measuring how many transactions have been sent from the account. It is used to prevent replay attacks.
*	`Gas` refers to how much [gas](../basics/gas-fees.md), which represents computational resources, Tx consumes. Gas is dependent on the transaction and is not precisely calculated until execution, but can be estimated by providing auto as the value for `Gas`.
*	`GasAdjustment` can be used to scale gas up in order to avoid underestimating. For example, users can specify their gas adjustment as 1.5 to use 1.5 times the estimated gas.
*	`GasPrices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, --gas-prices=0.025uatom, 0.025upho means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
*	`Fees` specifies how much in [fees](../basics/gas-fees.md) the user is willing to pay in total. Note that the user only needs to provide either `gas-prices` or `fees`, but not both, because they can be derived from each other.
*	`Simulate` instructs the application to ignore gas and simulate the transaction running without broadcasting.

### Request Handlers

Request handlers must be defined for both transaction and query requests. Handlers' arguments include a reference to the application's `codec` and the [`CLIContext`](../interfaces/query-lifecycle.md#clicontext) created in the user interaction.

Here is an example of a request handler for the nameservice module `buyNameReq` request (the same one shown above):

+++ https://github.com/cosmos/cosmos-sdk/blob/7f59723d889b69ca19966167f0b3a7fec7a39e53/x/bank/client/rest/tx.go#L21-L51

The request handler can be broken down as follows:

* **Parse Request:** The request handler first attempts to parse the request, and then run `Sanitize` and `ValidateBasic` on the underlying `BaseReq` to check the validity of the request. Next, it attempts to parse the arguments `Buyer` and `Amount` to the types `AccountAddress` and `Coins` respectively.
* **Message:** Then, a [message](./messages-and-queries.md) of the type `MsgBuyName` (defined by the module developer to trigger the state changes for this transaction) is created from the values and another sanity check, `ValidateBasic` is run on it.
* **Generate Transaction:** Finally, the HTTP `ResponseWriter`, application [`codec`](../core/encoding.md), [`CLIContext`](../interfaces/query-lifecycle.md#clicontext), request [`BaseReq`](../interfaces/rest.md#basereq), and message is passed to `WriteGenerateStdTxResponse` to further process the request.

To read more about how a transaction is generated, visit the transactions documentation [here](../core/transactions.md#transaction-generation).


### Register Routes

The application CLI entrypoint will have a `RegisterRoutes` function in its `main.go` file, which calls the `registerRoutes` functions of each module utilized by the application. Module developers need to implement `registerRoutes` for their modules so that applications are able to route messages and queries to their corresponding handlers and queriers.

The router used by the SDK is [Gorilla Mux](https://github.com/gorilla/mux). The router is initialized with the Gorilla Mux `NewRouter()` function. Then, the router's `HandleFunc` function can then be used to route urls with the defined request handlers and the HTTP method (e.g. "POST", "GET") as a route matcher. It is recommended to prefix every route with the name of the module to avoid collisions with other modules that have the same query or transaction names.

Here is a `registerRoutes` function with one query route example from the [nameservice tutorial](https://cosmos.network/docs/tutorial/rest.html):

``` go
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, storeName string) {
  // ResolveName Query
  r.HandleFunc(fmt.Sprintf("/%s/names/{%s}", storeName, restName), resolveNameHandler(cdc, cliCtx, storeName)).Methods("GET")
}
```

A few things to note:

* The router `r` has already been initialized by the application and is passed in here as an argument - this function is able to add on the nameservice module's routes onto any application's router. The application must also provide a [`CLIContext`](../interfaces/query-lifecycle.md#clicontext) that the querier will need to process user requests and the application [`codec`](../core/encoding.md) for encoding and decoding application-specific types.
* `"/%s/names/{%s}", storeName, restName` is the url for the HTTP request. `storeName` is the name of the module, `restName` is a variable provided by the user to specify what kind of query they are making.
* `resolveNameHandler` is the query request handler defined by the module developer. It also takes the application `codec` and `CLIContext` passed in from the user side, as well as the `storeName`.
* `"GET"` is the HTTP Request method. As to be expected, queries are typically GET requests. Transactions are typically POST and PUT requests.


## Next {hide}

Read about the recommended [module structure](./structure.md) {hide}
