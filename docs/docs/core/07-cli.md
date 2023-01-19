---
sidebar_position: 1
---

# Command-Line Interface

:::note Synopsis
This document describes how command-line interface (CLI) works on a high-level, for an [**application**](../basics/00-app-anatomy.md). A separate document for implementing a CLI for a Cosmos SDK [**module**](../building-modules/01-intro.md) can be found [here](../building-modules/09-module-interfaces.md#cli).
:::

## Command-Line Interface

### Example Command

There is no set way to create a CLI, but Cosmos SDK modules typically use the [Cobra Library](https://github.com/spf13/cobra). Building a CLI with Cobra entails defining commands, arguments, and flags. [**Commands**](#root-command) understand the actions users wish to take, such as `tx` for creating a transaction and `query` for querying the application. Each command can also have nested subcommands, necessary for naming the specific transaction type. Users also supply **Arguments**, such as account numbers to send coins to, and [**Flags**](#flags) to modify various aspects of the commands, such as gas prices or which node to broadcast to.

Here is an example of a command a user might enter to interact with the simapp CLI `simd` in order to send some tokens:

```bash
simd tx bank send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --gas auto --gas-prices <gasPrices>
```

The first four strings specify the command:

* The root command for the entire application `simd`.
* The subcommand `tx`, which contains all commands that let users create transactions.
* The subcommand `bank` to indicate which module to route the command to ([`x/bank`](../modules/bank/README.md) module in this case).
* The type of transaction `send`.

The next two strings are arguments: the `from_address` the user wishes to send from, the `to_address` of the recipient, and the `amount` they want to send. Finally, the last few strings of the command are optional flags to indicate how much the user is willing to pay in fees (calculated using the amount of gas used to execute the transaction and the gas prices provided by the user).

The CLI interacts with a [node](../core/03-node.md) to handle this command. The interface itself is defined in a `main.go` file.

### Building the CLI

The `main.go` file needs to have a `main()` function that creates a root command, to which all the application commands will be added as subcommands. The root command additionally handles:

* **setting configurations** by reading in configuration files (e.g. the Cosmos SDK config file).
* **adding any flags** to it, such as `--chain-id`.
* **instantiating the `codec`** by calling the application's `MakeCodec()` function (called `MakeTestEncodingConfig` in `simapp`). The [`codec`](../core/05-encoding.md) is used to encode and decode data structures for the application - stores can only persist `[]byte`s so the developer must define a serialization format for their data structures or use the default, Protobuf.
* **adding subcommand** for all the possible user interactions, including [transaction commands](#transaction-commands) and [query commands](#query-commands).

The `main()` function finally creates an executor and [execute](https://pkg.go.dev/github.com/spf13/cobra#Command.Execute) the root command. See an example of `main()` function from the `simapp` application:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/main.go#L12-L24
```

The rest of the document will detail what needs to be implemented for each step and include smaller portions of code from the `simapp` CLI files.

## Adding Commands to the CLI

Every application CLI first constructs a root command, then adds functionality by aggregating subcommands (often with further nested subcommands) using `rootCmd.AddCommand()`. The bulk of an application's unique capabilities lies in its transaction and query commands, called `TxCmd` and `QueryCmd` respectively.

### Root Command

The root command (called `rootCmd`) is what the user first types into the command line to indicate which application they wish to interact with. The string used to invoke the command (the "Use" field) is typically the name of the application suffixed with `-d`, e.g. `simd` or `gaiad`. The root command typically includes the following commands to support basic functionality in the application.

* **Status** command from the Cosmos SDK rpc client tools, which prints information about the status of the connected [`Node`](../core/03-node.md). The Status of a node includes `NodeInfo`,`SyncInfo` and `ValidatorInfo`.
* **Keys** [commands](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/client/keys) from the Cosmos SDK client tools, which includes a collection of subcommands for using the key functions in the Cosmos SDK crypto tools, including adding a new key and saving it to the keyring, listing all public keys stored in the keyring, and deleting a key. For example, users can type `simd keys add <name>` to add a new key and save an encrypted copy to the keyring, using the flag `--recover` to recover a private key from a seed phrase or the flag `--multisig` to group multiple keys together to create a multisig key. For full details on the `add` key command, see the code [here](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/client/keys/add.go). For more details about usage of `--keyring-backend` for storage of key credentials look at the [keyring docs](../run-node/00-keyring.md).
* **Server** commands from the Cosmos SDK server package. These commands are responsible for providing the mechanisms necessary to start an ABCI Tendermint application and provides the CLI framework (based on [cobra](https://github.com/spf13/cobra)) necessary to fully bootstrap an application. The package exposes two core functions: `StartCmd` and `ExportCmd` which creates commands to start the application and export state respectively.
Learn more [here](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/server).
* [**Transaction**](#transaction-commands) commands.
* [**Query**](#query-commands) commands.

Next is an example `rootCmd` function from the `simapp` application. It instantiates the root command, adds a [*persistent* flag](#flags) and `PreRun` function to be run before every execution, and adds all of the necessary subcommands.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L38-L92
```

`rootCmd` has a function called `initAppConfig()` which is useful for setting the application's custom configs.
By default app uses Tendermint app config template from Cosmos SDK, which can be over-written via `initAppConfig()`.
Here's an example code to override default `app.toml` template.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L106-L161
```

The `initAppConfig()` also allows overriding the default Cosmos SDK's [server config](https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/server/config/config.go#L235). One example is the `min-gas-prices` config, which defines the minimum gas prices a validator is willing to accept for processing a transaction. By default, the Cosmos SDK sets this parameter to `""` (empty string), which forces all validators to tweak their own `app.toml` and set a non-empty value, or else the node will halt on startup. This might not be the best UX for validators, so the chain developer can set a default `app.toml` value for validators inside this `initAppConfig()` function.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L126-L142
```

The root-level `status` and `keys` subcommands are common across most applications and do not interact with application state. The bulk of an application's functionality - what users can actually *do* with it - is enabled by its `tx` and `query` commands.

### Transaction Commands

[Transactions](./01-transactions.md) are objects wrapping [`Msg`s](../building-modules/02-messages-and-queries.md#messages) that trigger state changes. To enable the creation of transactions using the CLI interface, a function `txCommand` is generally added to the `rootCmd`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L177-L184
```

This `txCommand` function adds all the transaction available to end-users for the application. This typically includes:

* **Sign command** from the [`auth`](../modules/auth/README.md) module that signs messages in a transaction. To enable multisig, add the `auth` module's `MultiSign` command. Since every transaction requires some sort of signature in order to be valid, the signing command is necessary for every application.
* **Broadcast command** from the Cosmos SDK client tools, to broadcast transactions.
* **All [module transaction commands](../building-modules/09-module-interfaces.md#transaction-commands)** the application is dependent on, retrieved by using the [basic module manager's](../building-modules/01-module-manager.md#basic-manager) `AddTxCommands()` function.

Here is an example of a `txCommand` aggregating these subcommands from the `simapp` application:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L227-L251
```

### Query Commands

[**Queries**](../building-modules/02-messages-and-queries.md#queries) are objects that allow users to retrieve information about the application's state. To enable the creation of queries using the CLI interface, a function `queryCommand` is generally added to the `rootCmd`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L177-L184
```

This `queryCommand` function adds all the queries available to end-users for the application. This typically includes:

* **QueryTx** and/or other transaction query commands] from the `auth` module which allow the user to search for a transaction by inputting its hash, a list of tags, or a block height. These queries allow users to see if transactions have been included in a block.
* **Account command** from the `auth` module, which displays the state (e.g. account balance) of an account given an address.
* **Validator command** from the Cosmos SDK rpc client tools, which displays the validator set of a given height.
* **Block command** from the Cosmos SDK RPC client tools, which displays the block data for a given height.
* **All [module query commands](../building-modules/09-module-interfaces.md#query-commands)** the application is dependent on, retrieved by using the [basic module manager's](../building-modules/01-module-manager.md#basic-manager) `AddQueryCommands()` function.

Here is an example of a `queryCommand` aggregating subcommands from the `simapp` application:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L204-L225
```

## Flags

Flags are used to modify commands; developers can include them in a `flags.go` file with their CLI. Users can explicitly include them in commands or pre-configure them by inside their [`app.toml`](../run-node/02-interact-node.md#configuring-the-node-using-apptoml). Commonly pre-configured flags include the `--node` to connect to and `--chain-id` of the blockchain the user wishes to interact with.

A *persistent* flag (as opposed to a *local* flag) added to a command transcends all of its children: subcommands will inherit the configured values for these flags. Additionally, all flags have default values when they are added to commands; some toggle an option off but others are empty values that the user needs to override to create valid commands. A flag can be explicitly marked as *required* so that an error is automatically thrown if the user does not provide a value, but it is also acceptable to handle unexpected missing flags differently.

Flags are added to commands directly (generally in the [module's CLI file](../building-modules/09-module-interfaces.md#flags) where module commands are defined) and no flag except for the `rootCmd` persistent flags has to be added at application level. It is common to add a *persistent* flag for `--chain-id`, the unique identifier of the blockchain the application pertains to, to the root command. Adding this flag can be done in the `main()` function. Adding this flag makes sense as the chain ID should not be changing across commands in this application CLI.

## Environment variables

Each flag is bound to it's respecteve named environment variable. Then name of the environment variable consist of two parts - capital case `basename` followed by flag name of the flag. `-` must be substituted with `_`. For example flag `--home` for application with basename `GAIA` is bound to `GAIA_HOME`. It allows reducing the amount of flags typed for routine operations. For example instead of:

```shell
gaia --home=./ --node=<node address> --chain-id="testchain-1" --keyring-backend=test tx ... --from=<key name>
```

this will be more convenient:

```shell
# define env variables in .env, .envrc etc
GAIA_HOME=<path to home>
GAIA_NODE=<node address>
GAIA_CHAIN_ID="testchain-1"
GAIA_KEYRING_BACKEND="test"

# and later just use
gaia tx ... --from=<key name>
```

## Configurations

It is vital that the root command of an application uses `PersistentPreRun()` cobra command property for executing the command, so all child commands have access to the server and client contexts. These contexts are set as their default values initially and maybe modified, scoped to the command, in their respective `PersistentPreRun()` functions. Note that the `client.Context` is typically pre-populated with "default" values that may be useful for all commands to inherit and override if necessary.

Here is an example of an `PersistentPreRun()` function from `simapp`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.47.0-rc1/simapp/simd/cmd/root.go#L63-L86
```

The `SetCmdClientContextHandler` call reads persistent flags via `ReadPersistentCommandFlags` which creates a `client.Context` and sets that on the root command's `Context`.

The `InterceptConfigsPreRunHandler` call creates a viper literal, default `server.Context`, and a logger and sets that on the root command's `Context`. The `server.Context` will be modified and saved to disk via the internal `interceptConfigs` call, which either reads or creates a Tendermint configuration based on the home path provided. In addition, `interceptConfigs` also reads and loads the application configuration, `app.toml`, and binds that to the `server.Context` viper literal. This is vital so the application can get access to not only the CLI flags, but also to the application configuration values provided by this file.
