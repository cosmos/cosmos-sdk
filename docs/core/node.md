<!--
order: 4
synopsis: The main endpoint of an SDK application is the daemon client, otherwise known as the full-node client. The full-node runs the state-machine, starting from a genesis file. It connects to peers running the same client in order to receive and relay transactions, block proposals and signatures. The full-node is constituted of the application, defined with the Cosmos SDK, and of a consensus engine connected to the application via the ABCI. 
-->

# Node Client (Daemon)

## Pre-requisite Readings {hide}

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}

## `main` function

The full-node client of any SDK application is built by running a `main` function. The client is generally named by appending the `-d` suffix to the application name (e.g. `appd` for an application named `app`), and the `main` function is defined in a `./cmd/appd/main.go` file. Running this function creates an executable `.appd` that comes with a set of commands. For an app named `app`, the main command is [`appd start`](#start-command), which starts the full-node. 

In general, developers will implement the `main.go` function with the following structure:

- First, a [`codec`](./encoding.md) is instanciated for the application.
- Then, the `config` is retrieved and config parameters are set. This mainly involves setting the bech32 prefixes for [addresses and pubkeys](../basics/accounts.md#addresses-and-pubkeys).
	+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/config.go#L10-L21
- Using [cobra](https://github.com/spf13/cobra), the root command of the full-node client is created. After that, all the custom commands of the application are added using the `AddCommand()` method of `rootCmd`. 
- Add default server commands to `rootCmd` using the `server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)` method. These commands are separated from the ones added above since they are standard and defined at SDK level. They should be shared by all SDK-based applications. They include the most important command: the [`start` command](#start-command).
- Prepare and execute the `executor`.  
	+++ https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/libs/cli/setup.go#L75-L78

See an example of `main` function from the [`gaia`](https://github.com/cosmos/gaia) application:

+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/cmd/gaiad/main.go

## `start` command

The `start` command is defined in the `/server` folder of the Cosmos SDK. It is added to the root command of the full-node client in the [`main` function](#main-function) and called by the end-user to start their node:

```go
// For an example app named "app", the following command starts the full-node

appd start
```

As a reminder, the full-node is composed of three conceptual layers: the networking layer, the consensus layer and the application layer. The first two are generally bundled together in an entity called the consensus engine (Tendermint Core by default), while the third is the state-machine defined with the help of the Cosmos SDK. Currently, the Cosmos SDK uses Tendermint as the default consensus engine, meaning the start command is implemented to boot up a Tendermint node. 

The flow of the `start` command is pretty straightforward. First, it retrieves the `config` from the `context` in order to open the `db` (a [`leveldb`](https://github.com/syndtr/goleveldb) instance by default). This `db` contains the latest known state of the application (empty if the application is started from the first time. 

With the `db`, the `start` command creates a new instance of the application using an `appCreator` function:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/server/start.go#L144

Note that an `appCreator` is a function that fulfills the `AppCreator` signature. In practice, the [constructor the application](../basics/app-anatomy.md#constructor-function) is passed as the `appCreator`.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/server/constructors.go#L17-L25

Then, the instance of `app` is used to instanciate a new Tendermint node:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/server/start.go#L153-L163

The Tendermint node can be created with `app` because the latter satisfies the [`abci.Application` interface](https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/abci/types/application.go#L11-L26) (given that `app` extends [`baseapp`](./baseapp.md)). As part of the `NewNode` method, Tendermint makes sure that the height of the application (i.e. number of blocks since genesis) is equal to the height of the Tendermint node. The difference between these two heights should always be negative or null. If it is strictly negative, `NewNode` will replay blocks until the height of the application reaches the height of the Tendermint node. Finally, if the height of the application is `0`, the Tendermint node will call [`InitChain`](./baseapp.md#initchain) on the application to initialize the state from the genesis file. 

Once the Tendermint node is instanciated and in sync with the application, the node can be started:

```go
if err := tmNode.Start(); err != nil {
	return nil, err
}
```

Upon starting, the node will bootstrap its RPC and P2P server and start dialing peers. During handshake with its peers, if the node realizes they are ahead, it will query all the blocks sequentially in order to catch up. Then, it will wait for new block proposals and block signatures from validators in order to make progress. 

## Next {hide}

Learn about the [store](./store.md) {hide}