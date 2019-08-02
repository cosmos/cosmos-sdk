# Module Interfaces

## Prerequisites

* [Building Modules Intro](./intro.md)

## Synopsis

This document details how to build CLI and REST interfaces for a module. Examples from various SDK modules will be included.

- [CLI](#cli)
  + [Transaction Commands](#tx-commands)
  + [Query Commands](#query-commands)
- [REST](#rest)
  + [Request Types](#request-types)
  + [Request Handlers](#request-handlers)
  + [Register Routes](#register-routes)

## CLI

One of the main interfaces for an application is the [command-line interface](../interfaces/cli.md). This entrypoint created by the application developer will add commands from the application's modules to let end-users create [**messages**](./messages-and-queries.md) and [**queries**](./messages-and-queries.md).  The CLI files are typically found in the `./x/moduleName/client/cli` folder.

### Transaction Commands

[Transactions](../core/transactions.md) are created by users to wrap messages that trigger state changes when they get included in a valid block. Transaction commands typically have their own `tx.go` file in the module `./x/moduleName/client/cli` folder. The commands are specified in getter functions prefixed with `GetCmd` and include the name of the command. Here is an example from the nameservice tutorial:

```go
func GetCmdBuyName(cdc *codec.Codec) *cobra.Command {
  return &cobra.Command{
    Use: "buy-name [name] [amount]",
    Short: "bid for existing name or claim new name",
    Args: cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
      cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

      txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

      if err := cliCtx.EnsureAccountExists(); err != nil {
        return err
      }

      coins, err := sdk.ParseCoins(args[1])
      if err != nil {
        return err
      }

      msg := nameservice.NewMsgBuyName(args[0], coins, cliCtx.GetFromAddress())
      err = msg.ValidateBasic()
      if err != nil {
        return err
      }

      cliCtx.PrintResponse = true

      return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
    },
  }
}
```

This getter function creates the command for the Buy Name transaction. It does the following:

- **`codec`**. The getter function takes in an application [`codec`](../core/encoding.md) as an argument and returns a reference to the `cobra.command`. Since a module is intended to be used by any application, the `codec` must be provided.
- **Construct the command:** Read the [Cobra Documentation](https://github.com/spf13/cobra) for details on how to create commands.
  + **Use:** Specifies the format of a command-line entry users should type in order to invoke this command. In this case, the user uses `buy-name` as the name of the transaction command and provides the `name` the user wishes to buy and the `amount` the user is willing to pay.
  + **Args:** The number of arguments the user provides, in this case exactly two: `name` and `amount`.
  + **Short and Long:** A description for the function is provided here. A `Short` description is expected, and `Long` can be used to provide a more detailed description when a user uses the `--help` flag to ask for more information.
  + **RunE:** Defines a function that can return an error, called when the command is executed. Using `Run` would do the same thing, but would not allow for errors to be returned.
- **`RunE` Function Body:** The function should be specified as a `RunE` to allow for errors to be returned. This function encapsulates all of the logic to create a new transaction that is ready to be relayed to nodes.
  + The function should first initialize a [`TxBuilder`](../core/transactions.md#txbuilder) with the application `codec`'s `TxEncoder`, as well as a new [`CLIContext`](./query-lifecycle.md#clicontext) with the `codec` and `AccountDecoder`. These contexts contain all the information provided by the user and will be used to transfer this user-specific information between processes. To learn more about how contexts are used in a transaction, click [here](../core/transactions.md#transaction-generation).
  + If applicable, the command's arguments are parsed. Here, the `amount` given by the user is parsed into a denomination of `coins`.
  + If applicable, the `CLIContext` is used to retrieve any parameters such as the transaction originator's address to be used in the transaction. Here, the `from` address is retrieved by calling `cliCtx.getFromAddress()`.
  + A [message](./messages-and-queries.md) is created using all parameters parsed from the command arguments and `CLIContext`. The constructor function of the specific message type is called directly. It is good practice to call `ValidateBasic()` on the newly created message to run a sanity check and check for invalid arguments.
  + Depending on what the user wants, the transaction is either generated offline or signed and broadcasted to the preconfigured node using `GenerateOrBroadcastMsgs()`.
- **Flags.** Add any [flags](#flags) to the command. No flags were specified here, but all transaction commands have flags to provide additional information from the user (e.g. amount of fees they are willing to pay). These *persistent* [transaction flags](../interfaces/cli.md#flags) can be added to a higher-level command so that they apply to all transaction commands.


#### GetTxCmd

Finally, the module needs to have a `GetTxCmd()`, which aggregates all of the transaction commands of the module. Often, each command getter function has its own file in the module's `cli` folder, and a separate `tx.go` file contains `GetTxCmd()`. Application developers wishing to include the module's transactions will call this function to add them as subcommands in their CLI. Here is the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/67f6b021180c7ef0bcf25b6597a629aca27766b8/docs/spec/auth) `GetTxCmd()` function, which adds the `Sign` and `MultiSign` commands. An application using this module likely adds `auth` module commands to its root `TxCmd` command by calling `txCmd.AddCommand(authModuleClient.GetTxCmd())`.

```go
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetMultiSignCommand(cdc),
		GetSignCommand(cdc),
	)
	return txCmd
}
```

### Query Commands

[Queries](./messages-and-queries.md) allow users to gather information about the application or network state; they are routed by the application and processed by the module in which they are defined. Query commands typically have their own `query.go` file in the module `x/moduleName/client/cli` folder. Like transaction commands, they are specified in getter functions and have the prefix `GetCmdQuery`. Here is an example of a query command from the nameservice module:

```go
func GetCmdWhois(queryRoute string, cdc *codec.Codec) *cobra.Command {
 return &cobra.Command{
   Use: "whois [name]",
   Short: "Query whois info of name",
   Args: cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
     cliCtx := context.NewCLIContext().WithCodec(cdc)
     name := args[0]

     res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/whois/%s", queryRoute, name), nil)
     if err != nil {
       fmt.Printf("could not resolve whois - %s \n", string(name))
       return nil
     }

     var out nameservice.Whois
     cdc.MustUnmarshalJSON(res, &out)
     return cliCtx.PrintOutput(out)
   },
 }
}

```

This query returns the address that owns a particular name. The getter function does the following:

- **`codec` and `queryRoute`.** In addition to taking in the application `codec`, query command getters also take a `queryRoute` used to construct a path [Baseapp](../core/baseapp.md#query-routing) uses to route the query in the application.
- **Construct the command.** Read the [Cobra Documentation](https://github.com/spf13/cobra) and the [transaction command](#transaction-commands) example above for more information. The user must type `whois` and provide the `name` they are querying for as the only argument.
- **`RunE`.** The function should be specified as a `RunE` to allow for errors to be returned. This function encapsulates all of the logic to create a new query that is ready to be relayed to nodes.
  + The function should first initialize a new [`CLIContext`](./query-lifecycle.md#clicontext) with the application `codec`.
  + If applicable, the `CLIContext` is used to retrieve any parameters (e.g. the query originator's address to be used in the query) and marshal them with the query parameter type, in preparation to be relayed to a node. There are no `CLIContext` parameters in this case because the query does not involve any information about the user.
  + The `queryRoute` is used to construct a route Baseapp will use to route the query to the appropriate [querier](./querier.md). Module queries are `custom` type queries (some SDK modules have exceptions, such as `auth` and `gov` module queries).
  + The `CLIContext` query function relays the query to a node and retrieves the response.
  + The `codec` is used to nmarshal the response and the `CLIContext` is used to print the output back to the user.
- **Flags.** Add any [flags](#flags) to the command.

#### GetQueryCmd

Finally, the module also needs a `GetQueryCmd`, which aggregates all of the query commands of the module. Application developers wishing to include the module's queries will call this function to add them as subcommands in their CLI. Its structure is identical to the [`GetTxCmd`](#gettxcmd) command.

### Flags

[Flags](../interfaces/cli.md#flags) are entered by the user and allow for command customizations. Examples include the [fees](../core/accounts-fees.md) or gas prices users are willing to pay for their transactions.

The flags for a module are typically found in a `flags.go` file in the `./x/moduleName/client/cli` folder. Module developers can create a list of possible flags including the value type, default value, and a description displayed if the user uses a `help` command. In each transaction getter function, they can add flags to the commands and, optionally, mark flags as *required* so that an error is thrown if the user does not provide values for them.

For full details on flags, visit the [Cobra Documentation](https://github.com/spf13/cobra).

For example, the SDK `./client/flags` package includes a [`PostCommands()`](https://github.com/cosmos/cosmos-sdk/blob/master/client/flags/flags.go#L85-L116) function that adds necessary flags to transaction commands, such as the `from` flag to indicate which address the transaction originates from. Here is an example of how to add a flag using the `from` flag from this function.

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

Applications are typically required to support web services that use HTTP requests (e.g. a web wallet like [Lunie.io](lunie.io)). Thus, application developers will also use REST Routes to route HTTP requests to the application's modules; these routes will be used by service providers. The module developer's responsibility is to define the REST client by defining [routes](#register-routes) for all possible [requests](#request-types) and [handlers](#request-handlers) for each of them. It's up to the module developer how to organize the REST interface files; there is typically a `rest.go` file found in the module's `./x/moduleName/client/rest` folder.

### Request Types

Request types must be defined for all *transaction* requests. Conventionally, each request is named with the suffix `Req`, e.g. `SendReq` for a Send transaction. Each struct should include a base request [`baseReq`](../interfaces/rest.md#basereq), the name of the transaction, and all the arguments the user must provide for the transaction.

Here is an example of a request to buy a name from the [nameservice](https://cosmos.network/docs/tutorial/rest.html) module:

```go
type buyNameReq struct {
  BaseReq rest.BaseReq `json:"base_req"`
  Name string `json:"name"`
  Amount string `json:"amount"`
  Buyer string `json:"buyer"`
}
```
The `BaseReq` includes basic information that every request needs to have, similar to required flags in a CLI. All of these values, including `GasPrices` and `AccountNumber`, will be provided in the request body. The user will also need to specify the arguments `Name` and `Amount` fields in the body and `Buyer` will be provided by the user's address.

### Request Handlers

Request handlers must be defined for both transaction and query requests. Handlers' arguments include a reference to the application's `codec` and the [`CLIContext`](../interfaces/query-lifecycle.md#clicontext) created in the user interaction.

Here is an example of a request handler for the nameservice module `buyNameReq` request (the same one shown above):

```go
func buyNameHandler(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    var req buyNameReq

    if !rest.ReadRESTReq(w, r, cdc, &req) {
      rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
      return
    }

    baseReq := req.BaseReq.Sanitize()
    if !baseReq.ValidateBaic(w) {
      return
    }

    addr, err := sdk.AccAddressFromBech32(req.Buyer)
    if err != nil {
      rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
      return
    }

    coins, err := sdk.ParseCoins(req.Amount)
    if err != nil {
      rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
      return
    }

    //create message
    msg := nameservice.NewMsgBuyName(req.Name, coins, addr)
    err = msg.ValidateBasic()
    if err != nil {
      rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
      return
    }
    clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{msg})

  }
}
```

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


## Next

Read about the next topic in building modules.
