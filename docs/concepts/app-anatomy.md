# Anatomy of an SDK Application

## Pre-requisite reading

- [High-level overview of the architecture of an SDK application](../intro/sdk-app-architecture.md)
- [Cosmos SDK design overview](../intro/sdk-design.md)

## Synopsis

This document describes the core parts of a Cosmos SDK application. The placeholder name for this application will be `app`.

- [Node Client](#node-client)
- [Core Application File](#core-application-file)
- [Modules](#modules)
- [Intefaces](#interfaces)
- [Dependencies and Makefile](#dependencies-and-makefile)

The core parts listed above will generally translate to the following directory tree:

```
./app
├── cmd/
│   ├── appd
│   └── appcli
├── app.go
├── x/
│   ├── auth
│   └── bank
├── Gopkg.toml
└── Makefile
``` 

## Node Client (Daemon)

The Daemon, or Full-Node Client, is the core process of an SDK-based blockchain. Participants in the network run this process to initialize their state-machine, connect with other full-nodes and update their state-machine as new blocks come in. 

```
                ^  +-------------------------------+  ^
                |  |                               |  |
                |  |  State+machine = Application  |  |
                |  |                               |  |   Built with Cosmos SDK
                |  |            ^      +           |  |
                |  +----------- | ABCI | ----------+  v
                |  |            +      v           |  ^
                |  |                               |  |
Blockchain Node |  |           Consensus           |  |
                |  |                               |  |
                |  +-------------------------------+  |   Tendermint Core
                |  |                               |  |
                |  |           Networking          |  |
                |  |                               |  |
                v  +-------------------------------+  v
```
The blockchain full-node presents itself as a binary, generally suffixed by `-d` (e.g. `appd` for `app` or `gaiad` for the `gaia`) for "daemon". This binary is built by running a simple `main.go` function placed in `cmd/appd/`. This operation usually happens through the [Makefil](#dependencies-and-makefile).

To learn more about the `main.go` function, [click here](./node.md#main-function).

Once the main binary is built, the node can be started by running the `start` command. The core logic behind the `start` command is implemented in the SDK itself in the [`/server/start.go`](https://github.com/cosmos/cosmos-sdk/blob/master/server/start.go) file. The main [`start` command function](https://github.com/cosmos/cosmos-sdk/blob/master/server/start.go#L31) takes a [`context`](https://godoc.org/github.com/cosmos/cosmos-sdk/client/context) and [`appCreator`](#constructor-function-(`appCreator`)) as arguments. The `appCreator` is a constructor function for the SDK application, and is used in the starting process of the full-node. 

The `start` command function primarily does three things:

1- Create an instance of the state-machine defined in [`app.go`](#core-application-file) using the `appCreator`. 
2- Initialize the state-machine with the latest known state, extracted from the `db` stored in the `~/.appd/data` folder. At this point, the state-machine is at height `appBlockHeight`. 
3- Create and start a new Tendermint instance. Among other things, the node will perform a handshake with its peers. It will get the latest `blockHeight` from them, and replay blocks to sync to this height if it is greater than the local `appBlockHeight`. If `appBlockHeight` is `0`, the node is starting from genesis and Tendermint sends an `InitChain` message via the ABCI to the `app`, which triggers the [`InitChainer`](#initchainer).

To learn more about the `start` command, [click here](./node.md#start-command).

## Core Application File

In general, the core of the state-machine is defined in a file called `app.go`. It mainly contains the **type definition of the application** and functions to **create and initialize it**. 

### Type Definition of the Application

The first thing defined in `app.go` is the `type` of the application. It is generally comprised of the following parts:

- **A reference to [`baseapp`](./baseapp.md).** The custom application defined in `app.go` is a golang embedding of the `baseapp` type. `baseapp` implements most of the core logic for the application, including all the [ABCI methods](https://tendermint.com/docs/spec/abci/abci.html#overview) and the routing logic. When a transaction is relayed by Tendermint to the application, the latter uses `baseapp`'s methods to route them to the appropriate module. 
- **A list of store keys**. The [store](./store.md), which contains the entire state, is implemented as a multistore (i.e. a store of stores) in the Cosmos SDK. Each module uses one or multiple stores in the multistore to persist their part of the state. These stores can be accessed with specific keys that are declared in the `app` type. These keys, along with the `keepers`, are at the heart of the [object-capabilities model](../intro/ocap.md) of the Cosmos SDK.  
- **A list of module's `keepers`.** Each module defines an abstraction called `keeper`, which handles reads and writes for this module's store(s). The `keeper`'s methods of one module can be called from other modules (if authorized), which is why they are declared in the application's type. 
- **A reference to a `codec`.** The Cosmos SDK gives developers the freedom to choose the encoding framework for their application. The application's `codec` is used to serialize and deserialize data structures in order to store them, as stores can only persist `[]bytes`. The `codec` must be deterministic. Most SDK application use [amino](./amino.md) as their `codec`. 

You can see an example of application type definition [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L27-L43).

### Constructor Function

This function constructs a new application of the type defined above. It is [called](https://github.com/cosmos/cosmos-sdk/blob/master/server/start.go#L117) everytime the full-node is started with the `start` command. Here are the main actions performed by this function:

- Instanciate a new application with a reference to a `baseapp` instance, a codec and all the appropriate store keys.
- Instanciate all the `keepers` defined in the application's `type`.
- Initialize the application's [`routes`](./baseapp.md#routing) with the [`handlers`](#handler) of each one of the application's modules. When a transaction is relayed to the application by Tendermint via the ABCI, it is routed to the appropriate module's handler using the routes defined here. 
- Initialize the application's [query routes](./baseapp.md#query-routing) with the [`queriers`](#querier) of each of the application's modules. When a user query comes in, it is routed to the appropriate module using the query routes defined here. 
- Set the application's [`initChainer`](#initchainer) and mount the stores. 
- Return the application. 

Note that this function only creates an instance of the app, while the actual state is either carried over from the `~/.appd/data` folder if the node is restarted, or generated from the genesis file if the node is started for the first time. 

You can see an example of application constructor [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L46-L128).

### InitChainer

The `initChainer` is a function that initializes the state of the application from a [genesis file](./genesis.md) (i.e. token balances of genesis accounts). It is called when the application received the `InitChain` message from the Tendermint engine, which happens when the node is started at `appBlockHeight == 0` (i.e. on genesis). The application must set the `initChainer` in its constructor via the [`setInitChainer`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetInitChainer) method. 

You can see an example of an `initChainer` [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L137-L155).

### Register Codec

The `MakeCodec` function is the last important function of the `app.go` file. The goal of this function is to instanciate a codec `cdc` (e.g. [amino](./amino.md)) and calls the `RegisterCodec(*codec.Codec)` method of each module used within the application to register `cdc` to each module. 

In turn, the `RegisterCodec` function of each module register the custom interfaces and type structures of their respective module so that they can be marhsaled and unmarshaled. 

You can see an example of a `MakeCodec` [here](You can see an example of an `initChainer` [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/app.go#L189-L198).).

## Modules

Modules are the heart and soul of an SDK application. They can be considered as state-machines within the state-machine. When a transaction is relayed from the underlying Tendermint engine via the ABCI to the application, it is routed by `baseapp` to the appropriate module in order to be processed. This paradigm enables developers to easily build complex state-machines, as most of the modules they need often already exist. For developers, most of the work involved in building an SDK application revolves around building custom modules required by their application that do not exist, and integrating them with modules that do already exist into one coherent application. In the application directory, the standard practice is to store modules in the `x/` folder. 

To learn more about modules, [click here](./modules.md)

### Message Types

A message is a custom type defined by each module that implements the [`message`](https://github.com/cosmos/cosmos-sdk/blob/master/types/tx_msg.go#L8-L29) interface. Each `transaction` contains one or multiple `messages`. When a valid block of transactions is received by the full-node, Tendermint relays each one to the application via [`DeliverTx`](https://tendermint.com/docs/app-dev/abci-spec.html#delivertx). Upon receiving the transaction, the application first unmarshalls it. Then, it extracts the message(s) contained in the application. With the [`Type()`](https://github.com/cosmos/cosmos-sdk/blob/master/types/tx_msg.go#L16) method, `baseapp` is able to know which modules defines the message. It is then able to route it to the appropriate module's [handler](#handler) in order for the message to be processed. If the message is succesfully processed, the state is updated. 

Module developers create custom message types when they build their own module. The general practice is to prefix the type declaration of the message with `Msg`. For example, the message type `MsgSend` allows users to transfer tokens. It is processed by the handler of the `bank` module, which ultimately calls the `keeper` of the `auth` module in order to update the state.

To learn more about messages, [click here](./tx-msgs.md)

### Handler

The `handler` refers to the part of the module responsible for processing the message after it is routed by `baseapp`. `handler` functions of modules (except those of the `auth` module) are only executed if the transaction is relayed from Tendermint by the `DeliverTx` ABCI message. If the transaction is realyed by `CheckTx`, only stateless checks and fee-related (i.e. `auth` module-related) stateful checks are performed. To better understand the difference between `DeliverTx`and `CheckTx`, as well as the difference between stateful and stateless checks, click [here](./tx-lifecycle.md).

The handler of a module is generally defined in a file called `handler.go` and consists of:

- A **switch function** `NewHandler` to route the message to the appropriate handler function. This function returns a `handler` function, and is used in `app.go` to initialize the [application's router](./baseapp.md#routing). See an example of such a switch [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/handler.go#L10-L22).
- **One handler function for each message type defined by the module**. Developers write the message processing logic in these functions. This generally involves doing stateful checks to ensure the message is valid and calling [`keeper`](#keeper)'s methods to update the state. 

Handler functions return a result of type [`sdk.Result`](https://github.com/cosmos/cosmos-sdk/blob/master/types/result.go#L14-L37), which informs the application on wether the message was succesfully processed and.

To learn more about handlers, [click here](./handler.md).

### Keeper

`Keepers` are the gatekeepers of their module's store(s). To read or write in a module's store, it is mandatory to go through one of its `keeper`'s methods. This is ensured by the [object-capabilities](./ocap.md) model of the Cosmos SDK. Only objects that hold the key to a store can access it, and only the module's `keeper` should hold the key(s) to the module's store(s).

`Keepers` are generally defined in a file called `keeper.go`. It contains the `keeper`'s type definition and methods. 

The `keeper` type definition generally consists of:

- **Key(s)** to the module's store(s) in the multistore. 
- Reference to **other module's `keepers`**. Only needed if the `keeper` needs to access other module's store(s) (either to read or write from them).
- A reference to the application's **codec**. The `keeper` needs it to marshal structs before storing them, or to unmarhsal them when it retrieves them, because stores only accept `[]bytes` as value. 

The rest of the file defines the `keeper`'s methods, primarily getters and setters. You can check an example of a `keeper` implementation [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/keeper.go).

To learn more about `keepers`, [click here](./keeper.md).

### Querier 

`Queriers` are very similar to `handlers`, except they serve user queries to the state as opposed to processing transactions. A query is initiated from an [interface](#intefaces) by an end-user who provides a `queryRoute` and some `data`. The query is then routed to the correct application's `querier` by `baseapp`'s [`handleQueryCustom`](https://github.com/cosmos/cosmos-sdk/blob/master/baseapp/baseapp.go#L519-L556) method using `queryRoute`. 

The `Querier` of a module are defined in a file called `querier.go`, and consists of:

- A **switch function** `NewQuerier` to route the query to the appropriate `querier` function. This function returns a `querier` function, and is used in `app.go` to initialize the [application's query router](./baseapp.md#query-routing). See an example of such a switch [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/querier.go#L21-L34).
- - **One querier function for each data type defined by the module that needs to be queryable**. Developers write the query processing logic in these functions. This generally involves calling [`keeper`](#keeper)'s methods to query the state and marshalling it to JSON. See an example of `querier` functions [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/querier.go#L37-L101).

To learn more about `queriers`, [click here](./querier.md).

### Command-Line and REST Interfaces

Each module defines command-line commands and REST routes to be exposed to end-user via the [application's interfaces](#application-interfaces). This enables end-users to create messages of the types defined in the module, or to query the subset of the state managed by the module. 

#### CLI

Generally, the commands related to a module are defined in a folder called `client/cli` in the module's folder. The CLI divides commands in two category, transactions and queries, defined in `client/cli/tx.go` and `client/cli/query.go` respectively. Both build commands on top of the [Cobra Library](https://github.com/spf13/cobra):

- Transactions commands let users generate new transactions so that they can be included in a block and eventually update the state. One command should be created for each [message type](#message-types) defined in the module. The command calls the constructor of the message with the parameters provided by the end-user, and wraps it into a transaction. The SDK handles signing and the addition of other transaction metadata. See examples of transactions commands [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/client/cli/tx.go).
- Queries let users query the subset of the state defined by the module. Query commands forward queries to the [application's query router](./baseapp.md#query-routing), which routes them to the appropriate [querier](#querier) the `queryRoute` parameter supplied. See examples of query commands [here](https://github.com/cosmos/sdk-application-tutorial/blob/master/x/nameservice/client/cli/query.go).

To learn more about modules CLI, [click here](./module-interfaces.md#cli).

#### REST



## Application Interfaces

Developers build interfaces to let end-users interract with full-node clients. This means querying data from the full-node or creating and sending new transactions to be relayed by the full-node and eventually included in a block. 

### Command-Line Interface (CLI)

The main interface is the [Command-Line Interface](./interfaces.md#cli). The CLI of an SDK application is built from aggregating commands defined in each of the modules used by the application.

### REST Interface

A second important interface is the [REST interface](./interfaces.md#rest). It interract with the [Light Client Daemon](./node.md#lcd)

To learn more about interfaces, [click here](./interfaces.md)

## Dependencies and Makefile 

## Next

Learn more about the [Lifecycle of a transaction](./tx-lifecycle.md).