<!--
order: 11
-->

# Module Interfaces

This document details how to build CLI and REST interfaces for a module. Examples from various Cosmos SDK modules are included. {synopsis}

## Prerequisite Readings

- [Building Modules Intro](./intro.md) {prereq}

## CLI

One of the main interfaces for an application is the [command-line interface](../core/cli.md). This entrypoint adds commands from the application's modules enabling end-users to create [**messages**](./messages-and-queries.md#messages) wrapped in transactions and [**queries**](./messages-and-queries.md#queries). The CLI files are typically found in the module's `./client/cli` folder.

### Transaction Commands

 In order to create messages that trigger state changes, end-users must create [transactions](../core/transactions.md) that wrap and deliver the messages. A transaction command creates a transaction that includes one or more messages.

 Transaction commands typically have their own `tx.go` file that lives within the module's `./client/cli` folder. The commands are specified in getter functions and the name of the function should include the name of the command.

Here is an example from the `x/bank` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/bank/client/cli/tx.go#L28-L63

In the example, `NewSendTxCmd()` creates and returns the transaction command for a transaction that wraps and delivers `MsgSend`. `MsgSend` is the message used to send tokens from one account to another.

In general, the getter function does the following:

- **Constructs the command:** Read the [Cobra Documentation](https://godoc.org/github.com/spf13/cobra) for more detailed information on how to create commands.
    - **Use:** Specifies the format of the user input required to invoke the command. In the example above, `send` is the name of the transaction command and `[from_key_or_address]`, `[to_address]`, and `[amount]` are the arguments.
    - **Args:** The number of arguments the user provides. In this case, there are exactly three: `[from_key_or_address]`, `[to_address]`, and `[amount]`.
    - **Short and Long:** Descriptions for the command. A `Short` description is expected. A `Long` description can be used to provide additional information that is displayed when a user adds the `--help` flag.
    - **RunE:** Defines a function that can return an error. This is the function that is called when the command is executed. This function encapsulates all of the logic to create a new transaction.
        - The function typically starts by getting the `clientCtx`, which can be done with `client.GetClientTxContext(cmd)`. The `clientCtx` contains information relevant to transaction handling, including information about the user. In this example, the `clientCtx` is used to retrieve the address of the sender by calling `clientCtx.GetFromAddress()`.
        - If applicable, the command's arguments are parsed. In this example, the arguments `[to_address]` and `[amount]` are both parsed.
        - A [message](./messages-and-queries.md) is created using the parsed arguments and information from the `clientCtx`. The constructor function of the message type is called directly. In this case, `types.NewMsgSend(fromAddr, toAddr, amount)`. Its good practice to call [`msg.ValidateBasic()`](../basics/tx-lifecycle.md#ValidateBasic) and other validation methods before broadcasting the message.
        - Depending on what the user wants, the transaction is either generated offline or signed and broadcasted to the preconfigured node using `tx.GenerateOrBroadcastTxCLI(clientCtx, flags, msg)`.
- **Adds transaction flags:** All transaction commands must add a set of transaction [flags](#flags). The transaction flags are used to collect additional information from the user (e.g. the amount of fees the user is willing to pay). The transaction flags are added to the constructed command using `AddTxFlagsToCmd(cmd)`.
- **Returns the command:** Finally, the transaction command is returned.

Each module must implement `NewTxCmd()`, which aggregates all of the transaction commands of the module. Here is an example from the `x/bank` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/bank/client/cli/tx.go#L13-L26

Each module must also implement the `GetTxCmd()` method for `AppModuleBasic` that simply returns `NewTxCmd()`. This allows the root command to easily aggregate all of the transaction commands for each module. Here is an example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/bank/module.go#L75-L78

### Query Commands

[Queries](./messages-and-queries.md#queries) allow users to gather information about the application or network state; they are routed by the application and processed by the module in which they are defined. Query commands typically have their own `query.go` file in the module's `./client/cli` folder. Like transaction commands, they are specified in getter functions. Here is an example of a query command from the `x/auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/auth/client/cli/query.go#L75-L105

In the example, `GetAccountCmd()` creates and returns a query command that returns the state of an account based on the provided account address.

In general, the getter function does the following:

- **Constructs the command:** Read the [Cobra Documentation](https://godoc.org/github.com/spf13/cobra) for more detailed information on how to create commands.
    - **Use:** Specifies the format of the user input required to invoke the command. In the example above, `account` is the name of the query command and `[address]` is the argument.
    - **Args:** The number of arguments the user provides. In this case, there is exactly one: `[address]`.
    - **Short and Long:** Descriptions for the command. A `Short` description is expected. A `Long` description can be used to provide additional information that is displayed when a user adds the `--help` flag.
    - **RunE:** Defines a function that can return an error. This is the function that is called when the command is executed. This function encapsulates all of the logic to create a new query.
        - The function typically starts by getting the `clientCtx`, which can be done with `client.GetClientQueryContext(cmd)`. The `clientCtx` contains information relevant to query handling.
        - If applicable, the command's arguments are parsed. In this example, the argument `[address]` is parsed.
        - A new `queryClient` is initialized using `NewQueryClient(clientCtx)`. The `queryClient` is then used to call the appropriate [query](./messages-and-queries.md#grpc-queries).
        - The `clientCtx.PrintProto` method is used to format the `proto.Message` object so that the results can be printed back to the user.
- **Adds query flags:** All query commands must add a set of query [flags](#flags). The query flags are added to the constructed command using `AddQueryFlagsToCmd(cmd)`.
- **Returns the command:** Finally, the query command is returned.

Each module must implement `GetQueryCmd()`, which aggregates all of the query commands of the module. Here is an example from the `x/auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/auth/client/cli/query.go#L25-L42

Each module must also implement the `GetQueryCmd()` method for `AppModuleBasic` that returns the `GetQueryCmd()` function. This allows for the root command to easily aggregate all of the query commands for each module. Here is an example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/auth/module.go#L80-L83

### Flags

[Flags](../core/cli.md#flags) allow users to customize commands. `--fees` and `--gas-prices` are examples of flags that allow users to set the [fees](../basics/gas-fees.md) and gas prices for their transactions.

Flags that are specific to a module are typically created in a `flags.go` file in the module's `./client/cli` folder. When creating a flag, developers set the value type, the name of the flag, the default value, and a description about the flag. Developers also have the option to mark flags as _required_ so that an error is thrown if the user does not include a value for the flag.

Here is an example that adds the `--from` flag to a command:

```go
cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
```

In this example, the value of the flag is a `String`, the name of the flag is `from` (the value of the `FlagFrom` constant), the default value of the flag is `""`, and there is a description that will be displayed when a user adds `--help` to the command.

Here is an example that marks the `--from` flag as _required_:

```go
cmd.MarkFlagRequired(FlagFrom)
```

For more detailed information on creating flags, visit the [Cobra Documentation](https://github.com/spf13/cobra).

As mentioned in [transaction commands](#transaction-commands), there is a set of flags that all transaction commands must add. This is done with the `AddTxFlagsToCmd` method defined in the Cosmos SDK's `./client/flags` package.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/client/flags/flags.go#L94-L120

Since `AddTxFlagsToCmd(cmd *cobra.Command)` includes all of the basic flags required for a transaction command, module developers may choose not to add any of their own (specifying arguments instead may often be more appropriate).

Similarly, there is a `AddQueryFlagsToCmd(cmd *cobra.Command)` to add common flags to a module query command.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/client/flags/flags.go#L85-L92

## gRPC

[gRPC](https://grpc.io/) is a Remote Procedure Call (RPC) framework. RPC is the preferred way for external clients like wallets and exchanges to interact with a blockchain.

In addition to providing an ABCI query pathway, the Cosmos SDK provides a gRPC proxy server that routes gRPC query requests to ABCI query requests.

In order to do that, modules must implement `RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux)` on `AppModuleBasic` to wire the client gRPC requests to the correct handler inside the module.

Here's an example from the `x/auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/x/auth/module.go#L68-L73

## gRPC-gateway REST

Applications need to support web services that use HTTP requests (e.g. a web wallet like [Keplr](https://keplr.xyz)). [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) translates REST calls into gRPC calls, which might be useful for clients that do not use gRPC.

Modules that want to expose REST queries should add `google.api.http` annotations to their `rpc` methods, such as in the example below from the `x/auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-beta1/proto/cosmos/auth/v1beta1/query.proto#L13-L29

gRPC gateway is started in-process along with the application and Tendermint. It can be enabled or disabled by setting gRPC Configuration `enable` in [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml).

The Cosmos SDK provides a command for generating [Swagger](https://swagger.io/) documentation (`protoc-gen-swagger`). Setting `swagger` in [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml) defines if swagger documentation should be automatically registered.

## Next {hide}

Read about the recommended [module structure](./structure.md) {hide}
