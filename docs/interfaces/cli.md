<!--
order: 3
synopsis: "This document describes how to create a commmand-line interface (CLI) for an [**application**](../basics/app-anatomy.md). A separate document for implementing a CLI for an SDK [**module**](../building-modules/intro.md) can be found [here](#../building-modules/module-interfaces.md#cli)."
-->

# Command-Line Interface 

## Pre-requisite Readings {hide}

* [Lifecycle of a Query](./query-lifecycle.md) {prereq}

## Command-Line Interface

One of the main entrypoints of an application is the command-line interface. This entrypoint is created via a `main.go` file which compiles to a binary, conventionally placed in the application's `./cmd/cli` folder. The CLI for an application is typically be referred to as the name of the application suffixed with `-cli`, e.g. `appcli`. Here is where the interfaces docs lie in the directory from the [nameservice tutorial](https://cosmos.network/docs/tutorial).

### Example Command

There is no set way to create a CLI, but SDK modules typically use the [Cobra Library](https://github.com/spf13/cobra). Building a CLI with Cobra entails defining commands, arguments, and flags. [**Commands**](#commands) understand the actions users wish to take, such as `tx` for creating a transaction and `query` for querying the application. Each command can also have nested subcommands, necessary for naming the specific transaction type. Users also supply **Arguments**, such as account numbers to send coins to, and [**Flags**](#flags) to modify various aspects of the commands, such as gas prices or which node to broadcast to.

Here is an example of a command a user might enter to interact with the nameservice CLI `nscli` in order to buy a name:

```bash
nscli tx nameservice buy-name <name> <amount> --gas auto --gas-prices <gasPrices>
```

The first four strings specify the command: 

- The root command for the entire application `nscli`.
- The subcommand `tx`, which contains all commands that let users create transactions.
- The subcommand `nameservice` to indicate which module to route the command to (`nameservice` module in this case).
- The type of transaction `buy-name`. 

The next two strings are arguments: the `name` the user wishes to buy and the `amount` they want to pay for it. Finally, the last few strings of the command are flags to indicate how much the user is willing to pay in fees (calculated using the amount of gas used to execute the transaction and the gas prices provided by the user).

The CLI interacts with a [node](../core/node.md) (running `nsd`) to handle this command. The interface itself is defined in a `main.go` file.

### Building the CLI

The `main.go` file needs to have a `main()` function that does the following to run the command-line interface:

* **Instantiate the `codec`** by calling the application's `MakeCodec()` function. The [`codec`](../core/encoding.md) is used to code and encode data structures for the application - stores can only persist `[]byte`s so the developer must define a serialization format for their data structures or use the default, [Amino](../core/encoding.md#amino).
* **Configurations** are set by reading in configuration files (e.g. the sdk config file).
* **Create the root command** to which all the application commands will be added as subcommands and add any required flags to it, such as `--chain-id`.
* **Add subcommands** for all the possible user interactions, including [transaction commands](#transaction-commands) and [query commands](#query-commands).
* **Create an Executor** and [execute](https://godoc.org/github.com/spf13/cobra#Command.Execute) the root command.

See an example of `main()` function from the `nameservice` application:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L23-L66

The rest of the document will detail what needs to be implemented for each step and include smaller portions of code from the nameservice CLI `main.go` file.

## Adding Commands to the CLI

Every application CLI first constructs a root command, then adds functionality by aggregating subcommands (often with further nested subcommands) using `rootCmd.AddCommand()`. The bulk of an application's unique capabilities lies in its transaction and query commands, called `TxCmd` and `QueryCmd` respectively.

### Root Command

The root command (called `rootCmd`) is what the user first types into the command line to indicate which application they wish to interact with. The string used to invoke the command (the "Use" field) is typically the name of the application suffixed with `-cli`, e.g. `appcli`. The root command typically includes the following commands to support basic functionality in the application.

* **Status** command from the SDK rpc client tools, which prints information about the status of the connected [`Node`](../core/node.md). The Status of a node includes `NodeInfo`,`SyncInfo` and `ValidatorInfo`.
* **Config** [command](https://github.com/cosmos/cosmos-sdk/blob/master/client/config.go) from the SDK client tools, which allows the user to edit a `config.toml` file that sets values for [flags](#flags) such as `--chain-id` and which `--node` they wish to connect to.
The `config` command can be invoked by typing `appcli config` with optional arguments `<key> [value]` and a `--get` flag to query configurations or `--home` flag to create a new configuration.
* **Keys** [commands](https://github.com/cosmos/cosmos-sdk/blob/master/client/keys) from the SDK client tools, which includes a collection of subcommands for using the key functions in the SDK crypto tools, including adding a new key and saving it to disk, listing all public keys stored in the key manager, and deleting a key.
For example, users can type `appcli keys add <name>` to add a new key and save an encrypted copy to disk, using the flag `--recover` to recover a private key from a seed phrase or the flag `--multisig` to group multiple keys together to create a multisig key. For full details on the `add` key command, see the code [here](https://github.com/cosmos/cosmos-sdk/blob/master/client/keys/add.go).
* [**Transaction**](#transaction-commands) commands.
* [**Query**](#query-commands) commands.

Next is an example `main()` function from the `nameservice` application. It instantiates the root command, adds a [*persistent* flag](#flags) and `PreRun` function to be run before every execution, and adds all of the necessary subcommands.

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L23-L66

The root-level `status`, `config`, and `keys` subcommands are common across most applications and do not interact with application state. The bulk of an application's functionality - what users can actually *do* with it - is enabled by its transaction commands.

### Transaction Commands

[Transactions](#./transactions.md) are objects wrapping [messages](../building-modules/messages-and-queries.md#messages) that trigger state changes. To enable the creation of transactions using the CLI interface, a function `txCmd` is generally added to the `rootCmd`:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L51

This `txCmd` function adds all the transaction available to end-users for the application. This typically includes:

* **Sign command** from the [`auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/spec) module that signs messages in a transaction. To enable multisig, add the `auth` module's `MultiSign` command. Since every transaction requires some sort of signature in order to be valid, thithe signing command is necessary for every application.
* **Broadcast command** from the SDK client tools, to broadcast transactions.
* **Send command** from the [`bank`](https://github.com/cosmos/cosmos-sdk/tree/master/x/bank/spec) module, which is a transaction that allows accounts to send coins to one another, including gas and fees for transactions.
* **All [module transaction commands](../building-modules/module-interfaces.md#transaction-commands)** the application is dependent on, retrieved by using the [basic module manager's](../building-modules/module-manager.md#basic-manager) `AddTxCommands()` function.

Here is an example of a `txCmd` aggregating these subcommands from the `nameservice` application:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L97-L118


### Query Commands

[**Queries**](../building-modules/messages-and-queries.md#queries) are objects that allow users to retrieve information about the application's state. To enable the creation of transactions using the CLI interface, a function `txCmd` is generally added to the `rootCmd`:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L50

This `queryCmd` function adds all the queries available to end-users for the application. This typically includes:

* **QueryTx** and/or other transaction query commands] from the `auth` module which allow the user to search for a transaction by inputting its hash, a list of tags, or a block height. These queries allow users to see if transactions have been included in a block.
* **Account command** from the `auth` module, which displays the state (e.g. account balance) of an account given an address.
* **Validator command** from the SDK rpc client tools, which displays the validator set of a given height.
* **Block command** from the SDK rpc client tools, which displays the block data for a given height.
* **All [module query commands](../building-modules/module-interfaces.md#query-commands)** the application is dependent on, retrieved by using the [basic module manager's](../building-modules/module-manager.md#basic-manager) `AddQueryCommands()` function.

Here is an example of a `queryCmd` aggregating subcommands from the `nameservice` application:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L74-L95

## Flags

Flags are used to modify commands; developers can include them in a `flags.go` file with their CLI. Users can explicitly include them in commands or pre-configure them by entering a command in the format `appcli config <flag> <value>` into their command line. Commonly pre-configured flags include the `--node` to connect to and `--chain-id` of the blockchain the user wishes to interact with.

A *persistent* flag (as opposed to a _local_ flag) added to a command transcends all of its children: subcommands will inherit the configured values for these flags. Additionally, all flags have default values when they are added to commands; some toggle an option off but others are empty values that the user needs to override to create valid commands. A flag can be explicitly marked as _required_ so that an error is automatically thrown if the user does not provide a value, but it is also acceptable to handle unexpected missing flags differently.

Flags are added to commands directly (generally in the [module's CLI file](../building-modules/module-interfaces.md#flags) where module commands are defined) and no flag except for the `rootCmd` persistent flags has to be added at application level. It is common to add a _persistent_ flag for `--chain-id`, the unique identifier of the blockchain the application pertains to, to the root command. Adding this flag can be done in the `main()` function. Adding this flag makes sense as the chain ID should not be changing across commands in this application CLI. Here is an example from the `nameservice` application:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L41


## Configurations

The last function to define in `main.go` is `initConfig`, which does exactly what it sounds like - initialize configurations. To call this function, set it as a `PersistentPreRunE` function for the root command, so that it always executes before the main execution of the root command and any of its subcommands. `initConfig()` does the  following:

1. Read in the `config.toml` file. This same file is edited through `config` commands.
2. Use the [Viper](https://github.com/spf13/viper) to read in configurations from the file and set them.
3. Set any persistent flags defined by the user: `--chain-id`, `--encoding`, `--output`, etc.

Here is an example of an `initConfig()` function from the [nameservice tutorial CLI](https://cosmos.network/docs/tutorial/entrypoint.html#nscli):

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L120-L141

And an example of how to add `initConfig` as a `PersistentPreRunE` to the root command:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go#L42-L44

