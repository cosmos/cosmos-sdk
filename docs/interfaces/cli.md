# Command-Line Interface

## Prerequisites

* [Lifecycle of a Query](./query-lifecycle.md)

## Synopsis

This document describes how to create a commmand-line interface (CLI) for an application. A separate document for implementing a CLI for an SDK module can be found [here](#../building-modules/interfaces.md#cli).

- [Application CLI Components](#application-cli-components)
- [Commands](#commands)
- [Flags](#flags)
- [Configurations](#configurations)

## Application CLI Components

One of the main entrypoints of an application is the command-line interface. This entrypoint is created as a `main.go` file which compiles to a binary, conventionally placed in the application's `./cmd/cli` folder. The CLI for an application is typically be referred to as the name of the application suffixed with `-cli`, e.g. `appcli`. Here is where the interfaces docs lie in the directory from the [nameservice tutorial](https://cosmos.network/docs/tutorial)

### Cobra

There is no set way to create a CLI, but SDK modules all use the [Cobra Library](https://github.com/spf13/cobra). Building a CLI with Cobra entails defining commands, arguments, and flags. [**Commands**](#commands) understand the actions users wish to take, such as `tx` for creating a transaction and `query` for querying the application. Each command can also have nested subcommands, necessary for naming the specific transaction type. Users also supply **Arguments**, such as account numbers to send coins to, and [**Flags**](#flags) to modify various aspects of the commands, such as gas prices or which node to broadcast to.

Here is an example of a command a user might enter to interact with the nameservice CLI `nscli` in order to buy a name:

```bash
nscli tx nameservice buy-name <name> <amount> --gas auto --gas-prices <gasPrices>
```
The first four strings specify the command: the root command for the entire application `nscli`, the subcommand `tx`, the subcommand `nameservice` to indicate which module to route the command to, and the type of transaction `buy-name`. The next two strings are arguments: the `name` the user wishes to buy and the `amount` they want to pay for it. Finally, the last few strings of the command are flags to indicate how much the user is willing to pay in fees (calculated using the amount of gas used to execute the transaction and the gas prices provided by the user).

The CLI interacts with a node (running `nsd`) to handle this command.

### Main Function

The `main.go` file needs to have a `main()` function that does the following to run the command-line interface:

* **Instantiate the `codec`** by calling the application's `MakeCodec()` function. The [`codec`](../core/encoding.md) is used to code and encode data structures for the application - stores can only persist `[]byte`s so the developer must define a serialization format for their data structures or use the default, [Amino](../core/amino.md).
* **Configurations** are set by reading in configuration files (e.g. the sdk config file).
* **Create the root command** to which all the application commands will be added as subcommands and add any required flags to it, such as `--chain-id`.
* **Add subcommands** for all the possible user interactions, including [transaction commands](#transaction-commands) and [query commands](#query-commands).
* **Create an Executor** and execute the root command.

An example of the `main()` function for the [nameservice tutorial](https://cosmos.network/docs/tutorial) CLI can be found [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/cmd/nscli/main.go#L26-L67). The rest of the document will detail what needs to be implemented for each step and include smaller portions of code from the nameservice CLI `main.go` file.

## Commands

Every application CLI first constructs a root command, then adds functionality by aggregating subcommands (often with further nested subcommands) using `AddCommand()`. The bulk of an application's unique capabilities lies in its transaction and query commands, called `TxCmd` and `QueryCmd` respectively.

### Root Command

The root command (called `rootCmd`) is what the user first types into the command line to indicate which application they wish to interact with. The string used to invoke the command (the "Use" field) is typically the name of the application suffixed with `-cli`, e.g. `appcli`. The root command must include the following commands to support basic functionality in the application.

* **Status** command from the SDK rpc client tools, which prints information about the status of the connected [`Node`](,,/core/node.md). The Status of a node includes [`NodeInfo`](https://github.com/tendermint/tendermint/blob/master/p2p/node_info.go#L75-L92), `SyncInfo` and `ValidatorInfo`: this information includes the node ID, latest block hash, and the validator public key and voting power. Here is an example of what the `status command` outputs:
```json
{
  "jsonrpc": "2.0",
  "id": "",
  "result": {
    "node_info": {
    		"protocol_version": {
    			"p2p": "4",
    			"block": "7",
    			"app": "0"
    		},
    		"id": "53729852020041b956e86685e24394e0bee4373f",
    		"listen_addr": "10.0.2.15:26656",
    		"network": "test-chain-Y1OHx6",
    		"version": "0.24.0-2ce1abc2",
    		"channels": "4020212223303800",
    		"moniker": "ubuntu-xenial",
    		"other": {
    			"tx_index": "on",
    			"rpc_addr": "tcp://0.0.0.0:26657"
    		}
    	},
    	"sync_info": {
    		"latest_block_hash": "F51538DA498299F4C57AC8162AAFA0254CE08286",
    		"latest_app_hash": "0000000000000000",
    		"latest_block_height": "18",
    		"latest_block_time": "2018-09-17T11:42:19.149920551Z",
    		"catching_up": false
    	},
    	"validator_info": {
    		"address": "D9F56456D7C5793815D0E9AF07C3A355D0FC64FD",
    		"pub_key": {
    			"type": "tendermint/PubKeyEd25519",
    			"value": "wVxKNtEsJmR4vvh651LrVoRguPs+6yJJ9Bz174gw9DM="
    		},
    		"voting_power": "10"
    	}
    }
}
```
* **Config** [command](https://github.com/cosmos/cosmos-sdk/blob/master/client/config.go) from the SDK client tools, which allows the user to edit a `config.toml` file that sets values for [flags](#flags) such as `--chain-id` and which `--node` they wish to connect to.
The `config` command can be invoked by typing `appcli config` with optional arguments `<key> [value]` and a `--get` flag to query configurations or `--home` flag to create a new configuration.
* **Keys** [commands](https://github.com/cosmos/cosmos-sdk/blob/master/client/keys) from the SDK client tools, which includes a collection of subcommands for using the key functions in the SDK crypto tools, including adding a new key and saving it to disk, listing all public keys stored in the key manager, and deleting a key.
For example, users can type `appcli keys add <name>` to add a new key and save an encrypted copy to disk, using the flag `--recover` to recover a private key from a seed phrase or the flag `--multisig` to group multiple keys together to create a multisig key. For full details on the `add` key command, see the code [here](https://github.com/cosmos/cosmos-sdk/blob/master/client/keys/add.go).
* [**Transaction**](#transaction-commands) commands.
* [**Query**](#query-commands) commands.

Here is an example from the [nameservice tutorial](https://cosmos.network/docs/tutorial) CLI's `main()` function. It instantiates the root command, adds a [*persistent* flag](#flags) and `PreRun` function to be run before every execution, and adds all of the necessary subcommands.


```go
rootCmd := &cobra.Command{
  Use:   "nscli",
  Short: "nameservice Client",
}
rootCmd.AddCommand(
  rpc.StatusCommand(),
  client.ConfigCmd(defaultCLIHome),
  queryCmd(cdc, mc),
  txCmd(cdc, mc),
  client.LineBreak,
  lcd.ServeCommand(cdc, registerRoutes),
  client.LineBreak,
  keys.Commands(),
  client.LineBreak,
)
```

All of these things are done within the `main()` function. At the end of the `main()` function, it is necessary to create an `executor` and `Execute()` the root command in the `main()` function:

```go
executor := cli.PrepareMainCmd(rootCmd, "NS", defaultCLIHome)
err := executor.Execute()
```

### Transaction Commands

[Transactions](#./transactions.md) are objects wrapping [messages](../building-modules/messages-and-queries.md) that trigger state changes. To enable the creation of transactions using the CLI interface, `TxCmd` needs to add the following commands:

* **Sign** [command](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/client/cli/tx_sign.go#L30-L83) from the [`auth`](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/auth) module, the command that signs messages in a transaction. To enable multisig, add the `auth` module [`MultiSign`](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/client/cli/tx_multisign.go#L26-L151) command. Since every transaction requires some sort of signature in order to be valid, this command is necessary for every application.
* **Broadcast** [command](https://github.com/cosmos/cosmos-sdk/blob/master/client/context/broadcast.go) from the SDK client tools, which broadcasts transactions.
* **Send** [command](https://github.com/cosmos/cosmos-sdk/blob/master/x/bank/client/cli/tx.go#L31-L60) from the [`bank`](https://github.com/cosmos/cosmos-sdk/blob/master/docs/spec/bank) module, which is a transaction that allows accounts to send coins to one another, including gas and fees for transactions.
* All [module transaction commands](../building-modules/interfaces.md) the application is dependent on, retrieved by calling [`GetTxCmd()`](../building-modules/interfaces.md#GetTxCmd) on all the modules or using the Module Manager's [`AddTxCommands()`](../building-modules/module-manager.md) function.

Here is an example of a `TxCmd` aggregating these subcommands from the [nameservice tutorial](https://cosmos.network/docs/tutorial).

```go
func txCmd(cdc *amino.Codec, mc []sdk.ModuleClients) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		tx.GetBroadcastCommand(cdc),
		client.LineBreak,
	)

	for _, m := range mc {
		txCmd.AddCommand(m.GetTxCmd())
	}

	return txCmd
}
```


### Query Commands

[**Queries**](../building-modules/messages-and-queries.md#queries) are objects that allow users to retrieve information about the application's state. To enable basic queries, `QueryCmd` needs to add the following commands:

* **QueryTx** and/or other transaction [query commands](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/client/cli/query.go) from the `auth` module which allow the user to search for a transaction by inputting its hash, a list of tags, or a block height. These queries allow users to see if transactions have been included in a block.
* **Account** [command](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/client/cli/query.go#L45-L73) from the `auth` module, which displays the state (e.g. account balance) of an account given an address.
* **Validator** [command](https://github.com/cosmos/cosmos-sdk/blob/master/client/rpc/validators.go) from the SDK rpc client tools, which displays the validator set of a given height.
* **Block** [command](https://github.com/cosmos/cosmos-sdk/blob/master/client/rpc/block.go) from the SDK rpc client tools, which displays the block data for a given height.
* All [module query commands](../building-modules/interfaces.md) the application is dependent on, retrieved by calling [`GetQueryCmd()`](../building-modules/interfaces.md#GetqueryCmd) on all the modules or using the Module Manager's `AddQueryCommands()` function.

Here is an example of a `QueryCmd` aggregating subcommands, also from the nameservice tutorial (it is structurally identical to `TxCmd`):

```go
func queryCmd(cdc *amino.Codec, mc []sdk.ModuleClients) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	queryCmd.AddCommand(
		rpc.ValidatorCommand(cdc),
		rpc.BlockCommand(),
		tx.SearchTxCmd(cdc),
		tx.QueryTxCmd(cdc),
		client.LineBreak,
		authcmd.GetAccountCmd(storeAcc, cdc),
	)

	for _, m := range mc {
		queryCmd.AddCommand(m.GetQueryCmd())
	}

	return queryCmd
}
```

## Flags

Flags are used to modify commands. Users can explicitly include them in commands or pre-configure them by entering a command in the format `appcli config <flag> <value>` into their command line. Commonly pre-configured flags include the `--node` to connect to and `--chain-id` of the blockchain the user wishes to interact with.

A *persistent* flag (as opposed to a _local_ flag) added to a command transcends all of its children: subcommands will inherit the configured values for these flags. Additionally, all flags have default values when they are added to commands; some toggle an option off but others are empty values that the user needs to override to create valid commands. A flag can be explicitly marked as _required_ so that an error is automatically thrown if the user does not provide a value, but it is also acceptable to handle unexpected missing flags differently.

Every flag has a name the user types to use the flag. For example, the commonly used `--from` flag is declared as a constant in the SDK [flags](https://github.com/cosmos/cosmos-sdk/blob/master/client/flags/flags.go) file:

```go
const FlagFrom = "from"
```

The flag can be added to a command `cmd`, adding a default value and description:

```go
cmd.Flags().String(FlagFrom, "", "Name or address of private key with which to sign")
```

The SDK client package includes a list of [flags](https://github.com/cosmos/cosmos-sdk/blob/master/client/flags/flags.go) that are commonly used across existing commands.

### Root Command Flags

It is common to add a _persistent_ flag for `--chain-id`, the unique identifier of the blockchain the application pertains to, to the root command. Adding this flag makes sense as the chain ID should not be changing across commands in this application CLI. Here is what that looks like:

```go
rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
```

### Transaction and Query Flags

To create a transaction, users enter a `tx` command and provide several flags. The SDK `./client/flags` package [`PostCommands()`](https://github.com/cosmos/cosmos-sdk/blob/master/client/flags/flags.go#L85-L116) function adds a set of basic flags to every transaction command. For queries, the [`GetCommand()`](https://github.com/cosmos/cosmos-sdk/blob/master/client/flags/flags.go#L67-L82) function adds basic flags to query commands, such as the block `--height` to query from. 

For example, the following command creates a transaction to send 1000uatom from `sender-address` to `recipient-address`. The user is willing to pay 0.025uatom per unit gas but wants the transaction to be only generated offline (i.e. not broadcasted) and written, in JSON format, to the file `myUnsignedTx.json`.

```bash
appcli tx send <recipientAddress> 1000uatom --from <senderAddress> --gas auto -gas-prices 0.025uatom --generate-only > myUnsignedTx.json
```

Here are the flags used:

* `--from` indicates which [account](../core/accounts-fees.md) the transaction originates from. This account is used to sign the transaction.
* `--gas` refers to how much [gas](../core/gas.md), which represents computational resources, Tx consumes. Gas is dependent on the computational needs of the transaction and is not precisely calculated until execution, but can be estimated by providing auto as the value for --gas.
* `--gas-prices` specifies how much the user is willing pay per unit of gas, which can be one or multiple denominations of tokens. For example, --gas-prices=0.025uatom, 0.025upho means the user is willing to pay 0.025uatom AND 0.025upho per unit of gas.
* `--generate-only` (optional) instructs the application to simply generate the unsigned transaction and output or write to a file. Without this flag, the transaction is created, signed, and broadcasted all in one command.


## Configurations

The last function to define is, `initConfig`, which does exactly what it sounds like - initial configurations. To call this function, set it as a `PersistentPreRunE` function for the root command, so that it always executes before the main execution of the root command and any of its subcommands. `initConfig()` does the  following:

1. Read in the `config.toml` file. This same file is edited through `config` commands.
2. Use the [Viper](https://github.com/spf13/viper) to read in configurations from the file and set them.
3. Set any persistent flags defined by the user: `--chain-id`, `--encoding`, `--output`, etc.

Here is an example of an `initConfig()` function from the nameservice tutorial CLI:

```go
func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}
```

Here is an example of how to add `initConfig` as a `PersistentPreRunE` to the root command:

```go
rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
  return initConfig(rootCmd)
}
```

## Next

Read about how to build a module CLI [here](./module-interfaces#cli)
