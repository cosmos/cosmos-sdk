<!--
order: 1
synopsis: "This document describes the core parts of a Cosmos SDK application. Throughout the document, a placeholder application named `app` will be used."
-->

# Anatomy of an SDK Application 

## Node Client

The Daemon, or [Full-Node Client](../core/node.md), is the core process of an SDK-based blockchain. Participants in the network run this process to initialize their state-machine, connect with other full-nodes and update their state-machine as new blocks come in.

```
                ^  +-------------------------------+  ^
                |  |                               |  |
                |  |  State-machine = Application  |  |
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

The blockchain full-node presents itself as a binary, generally suffixed by `-d` for "daemon" (e.g. `appd` for `app` or `gaiad` for `gaia`). This binary is built by running a simple [`main.go`](../core/node.md#main-function) function placed in `./cmd/appd/`. This operation usually happens through the [Makefile](#dependencies-and-makefile).

Once the main binary is built, the node can be started by running the [`start` command](../core/node.md#start-command). This command function primarily does three things:

1. Create an instance of the state-machine defined in [`app.go`](#core-application-file).
2. Initialize the state-machine with the latest known state, extracted from the `db` stored in the `~/.appd/data` folder. At this point, the state-machine is at height `appBlockHeight`.
3. Create and start a new Tendermint instance. Among other things, the node will perform a handshake with its peers. It will get the latest `blockHeight` from them, and replay blocks to sync to this height if it is greater than the local `appBlockHeight`. If `appBlockHeight` is `0`, the node is starting from genesis and Tendermint sends an `InitChain` message via the ABCI to the `app`, which triggers the [`InitChainer`](#initchainer).

## Core Application File

In general, the core of the state-machine is defined in a file called `app.go`. It mainly contains the **type definition of the application** and functions to **create and initialize it**.

### Type Definition of the Application

The first thing defined in `app.go` is the `type` of the application. It is generally comprised of the following parts:

- **A reference to [`baseapp`](../core/baseapp.md).** The custom application defined in `app.go` is an extension of `baseapp`. When a transaction is relayed by Tendermint to the application, `app` uses `baseapp`'s methods to route them to the appropriate module. `baseapp` implements most of the core logic for the application, including all the [ABCI methods](https://tendermint.com/docs/spec/abci/abci.html#overview) and the [routing logic](../core/baseapp.md#routing). 
- **A list of store keys**. The [store](../core/store.md), which contains the entire state, is implemented as a [`multistore`](../core/store.md#multistore) (i.e. a store of stores) in the Cosmos SDK. Each module uses one or multiple stores in the multistore to persist their part of the state. These stores can be accessed with specific keys that are declared in the `app` type. These keys, along with the `keepers`, are at the heart of the [object-capabilities model](../core/ocap.md) of the Cosmos SDK.  
- **A list of module's `keeper`s.** Each module defines an abstraction called [`keeper`](../building-modules/keeper.md), which handles reads and writes for this module's store(s). The `keeper`'s methods of one module can be called from other modules (if authorized), which is why they are declared in the application's type and exported as interfaces to other modules so that the latter can only access the authorized functions. 
- **A reference to a [`codec`](../core/encoding.md).** The application's `codec` is used to serialize and deserialize data structures in order to store them, as stores can only persist `[]bytes`. The `codec` must be deterministic. The default codec is [amino](../core/encoding.md).
- **A reference to a [module manager](../building-modules/module-manager.md#manager)** and a [basic module manager](../building-modules/module-manager.md#basicmanager). The module manager is an object that contains a list of the application's module. It facilitates operations related to these modules, like registering [`routes`](../core/baseapp.md#routing), [query routes](../core/baseapp.md#query-routing) or setting the order of execution between modules for various functions like [`InitChainer`](#initchainer), [`BeginBlocker` and `EndBlocker`](#beginblocker-and-endblocker).

See an example of application type definition from [`gaia`](https://github.com/cosmos/gaia)

+++ https://github.com/cosmos/gaia/blob/5bc422e6868d04747e50b467e8eeb31ae2fe98a3/app/app.go#L87-L115

### Constructor Function

This function constructs a new application of the type defined in the section above. It must fulfill the `AppCreator` signature in order to be used in the [`start` command](../core/node.md#start-command) of the application's daemon command. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/server/constructors.go#L20

Here are the main actions performed by this function:

- Instantiate a new [`codec`](../core/encoding.md) and initialize the `codec` of each of the application's module using the [basic manager](../building-modules/module-manager.md#basicmanager)
- Instantiate a new application with a reference to a `baseapp` instance, a codec and all the appropriate store keys.
- Instantiate all the [`keeper`s](#keeper) defined in the application's `type` using the `NewKeeper` function of each of the application's modules. Note that `keepers` must be instantiated in the correct order, as the `NewKeeper` of one module might require a reference to another module's `keeper`. 
- Instantiate the application's [module manager](../building-modules/module-manager.md#manager) with the [`AppModule`](#application-module-interface) object of each of the application's modules. 
- With the module manager, initialize the application's [`routes`](../core/baseapp.md#routing) and [query routes](../core/baseapp.md#query-routing). When a transaction is relayed to the application by Tendermint via the ABCI, it is routed to the appropriate module's [`handler`](#handler) using the routes defined here. Likewise, when a query is received by the application, it is routed to the appropriate module's [`querier`](#querier) using the query routes defined here. 
- With the module manager, register the [application's modules' invariants](../building-modules/invariants.md). Invariants are variables (e.g. total supply of a token) that are evaluated at the end of each block. The process of checking invariants is done via a special module called the [`InvariantsRegistry`](../building-modules/invariants.md#invariant-registry). The value of the invariant should be equal to a predicted value defined in the module. Should the value be different than the predicted one, special logic defined in the invariant registry will be triggered (usually the chain is halted). This is useful to make sure no critical bug goes unnoticed and produces long-lasting effects that would be hard to fix. 
- With the module manager, set the order of execution between the `InitGenesis`, `BegingBlocker` and `EndBlocker` functions of each of the [application's modules](#application-module-interface). Note that not all modules implement these functions.  
- Set the remainer of application's parameters:
    + [`InitChainer`](#initchainer): used to initialize the application when it is first started.
    + [`BeginBlocker`, `EndBlocker`](#beginblocker-and-endlbocker): called at the beginning and the end of every block).
    + [`anteHandler`](../core/baseapp.md#antehandler): used to handle fees and signature verification.  
- Mount the stores. 
- Return the application. 

Note that this function only creates an instance of the app, while the actual state is either carried over from the `~/.appd/data` folder if the node is restarted, or generated from the genesis file if the node is started for the first time.

See an example of application constructor from [`gaia`](https://github.com/cosmos/gaia):

+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L110-L222

### InitChainer

The `InitChainer` is a function that initializes the state of the application from a genesis file (i.e. token balances of genesis accounts). It is called when the application receives the `InitChain` message from the Tendermint engine, which happens when the node is started at `appBlockHeight == 0` (i.e. on genesis). The application must set the `InitChainer` in its [constructor](#constructor-function) via the [`SetInitChainer`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetInitChainer) method. 

In general, the `InitChainer` is mostly composed of the [`InitGenesis`](../building-modules/genesis.md#initgenesis) function of each of the application's modules. This is done by calling the `InitGenesis` function of the module manager, which in turn will call the `InitGenesis` function of each of the modules it contains. Note that the order in which the modules' `InitGenesis` functions must be called has to be set in the module manager using the [module manager's](../building-modules/module-manager.md) `SetOrderInitGenesis` method. This is done in the [application's constructor](#application-constructor), and the `SetOrderInitGenesis` has to be called before the `SetInitChainer`.

See an example of an `InitChainer` from [`gaia`](https://github.com/cosmos/gaia):
    
+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L235-L239

### BeginBlocker and EndBlocker

The SDK offers developers the possibility to implement automatic execution of code as part of their application. This is implemented through two function called `BeginBlocker` and `EndBlocker`. They are called when the application receives respectively the `BeginBlock` and `EndBlock` messages from the Tendermint engine, which happens at the beginning and at the end of each block. The application must set the `BeginBlocker` and `EndBlocker` in its [constructor](#constructor-function) via the [`SetBeginBlocker`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetBeginBlocker) and [`SetEndBlocker`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetEndBlocker) methods.

In general, the `BeginBlocker` and `EndBlocker` functions are mostly composed of the [`BeginBlock` and `EndBlock`](../building-modules/beginblock-endblock.md) functions of each of the application's modules. This is done by calling the `BeginBlock` and `EndBlock` functions of the module manager, which in turn will call the `BeginBLock` and `EndBlock` functions of each of the modules it contains. Note that the order in which the modules' `BegingBlock` and `EndBlock` functions must be called has to be set in the module manager using the `SetOrderBeginBlock` and `SetOrderEndBlock` methods respectively. This is done via the [module manager](../building-modules/module-manager.md) in the [application's constructor](#application-constructor), and the `SetOrderBeginBlock` and `SetOrderEndBlock` methods have to be called before the `SetBeginBlocker` and `SetEndBlocker` functions.

As a sidenote, it is important to remember that application-specific blockchains are deterministic. Developers must be careful not to introduce non-determinism in `BeginBlocker` or `EndBlocker`, and must also be careful not to make them too computationally expensive, as [gas](./gas-fees.md) does not constrain the cost of `BeginBlocker` and `EndBlocker` execution. 

See an example of `BeginBlocker` and `EndBlocker` functions from [`gaia`](https://github.com/cosmos/gaia)

+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L224-L232

### Register Codec

The `MakeCodec` function is the last important function of the `app.go` file. The goal of this function is to instantiate a [codec `cdc`](../core/encoding.md) (e.g. amino) initialize the codec of the SDK and each of the application's modules using the `RegisterCodec` function.

To register the application's modules, the `MakeCodec` function calls `RegisterCodec` on `ModuleBasics`. `ModuleBasics` is a [basic manager](../building-modules/module-manager.md#basicmanager) which lists all of the application's modules. It is instanciated in the `init()` function, and only serves to easily register non-dependant elements of application's modules (such as codec). To learn more about the basic module manager, click [here](../building-modules/module-manager.md#basicmanager).

See an example of a `MakeCodec` from [`gaia`](https://github.com/cosmos/gaia):
    
+++ https://github.com/cosmos/gaia/blob/f41a660cdd5bea173139965ade55bd25d1ee3429/app/app.go#L64-L70

## Modules

[Modules](../building-modules/intro.md) are the heart and soul of SDK applications. They can be considered as state-machines within the state-machine. When a transaction is relayed from the underlying Tendermint engine via the ABCI to the application, it is routed by [`baseapp`](../core/baseapp.md) to the appropriate module in order to be processed. This paradigm enables developers to easily build complex state-machines, as most of the modules they need often already exist. For developers, most of the work involved in building an SDK application revolves around building custom modules required by their application that do not exist yet, and integrating them with modules that do already exist into one coherent application. In the application directory, the standard practice is to store modules in the `x/` folder (not to be confused with the SDK's `x/` folder, which contains already-built modules). 

### Application Module Interface

Modules must implement [interfaces](../building-modules/module-manager.md#application-module-interfaces) defined in the Cosmos SDK, [`AppModuleBasic`](../building-modules/module-manager.md#appmodulebasic) and [`AppModule`](../building-modules/module-manager.md#appmodule). The former implements basic non-dependant elements of the module, such as the `codec`, while the latter handles the bulk of the module methods (including methods that require references to other modules' `keeper`s). Both the `AppModule` and `AppModuleBasic` types are defined in a file called `./module.go`. 

`AppModule` exposes a collection of useful methods on the module that facilitates the composition of modules into a coherent application. These methods are are called from the `module manager`(../building-modules/module-manager.md#manager), which manages the application's collection of modules. 

### Message Types

[`Message`s](../building-modules/messages-and-queries.md#messages) are objects defined by each module that implement the [`message`](../building-modules/messages-and-queries.md#messages) interface. Each [`transaction`](../core/transactions.md) contains one or multiple `messages`. 

When a valid block of transactions is received by the full-node, Tendermint relays each one to the application via [`DeliverTx`](https://tendermint.com/docs/app-dev/abci-spec.html#delivertx). Then, the application handles the transaction:

1. Upon receiving the transaction, the application first unmarshalls it from `[]bytes`.
2. Then, it verifies a few things about the transaction like [fee payment and signatures](#gas-fees.md#antehandler) before extracting the message(s) contained in the transaction. 
3. With the `Type()` method of the `message`, `baseapp` is able to route it to the appropriate module's [`handler`](#handler) in order for it to be processed. 
4. If the message is successfully processed, the state is updated. 

For a more detailed look at a transaction lifecycle, click [here](./tx-lifecycle.md).

Module developers create custom message types when they build their own module. The general practice is to prefix the type declaration of the message with `Msg`. For example, the message type `MsgSend` allows users to transfer tokens:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/bank/internal/types/msgs.go#L10-L15

It is processed by the `handler` of the `bank` module, which ultimately calls the `keeper` of the `auth` module in order to update the state.

### Handler

The [`handler`](../building-modules/handler.md) refers to the part of the module responsible for processing the `message` after it is routed by `baseapp`. `handler` functions of modules are only executed if the transaction is relayed from Tendermint by the `DeliverTx` ABCI message. If the transaction is relayed by `CheckTx`, only stateless checks and fee-related stateful checks are performed. To better understand the difference between `DeliverTx`and `CheckTx`, as well as the difference between stateful and stateless checks, click [here](./tx-lifecycle.md).

The `handler` of a module is generally defined in a file called `handler.go` and consists of:

- A **switch function** `NewHandler` to route the message to the appropriate `handler` function. This function returns a `handler` function, and is registered in the [`AppModule`](#application-module-interface) to be used in the application's module manager to initialize the [application's router](../core/baseapp.md#routing). Next is an example of such a switch from the [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice)

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/internal/keeper/querier.go#L19-L32

- **One handler function for each message type defined by the module**. Developers write the message processing logic in these functions. This generally involves doing stateful checks to ensure the message is valid and calling [`keeper`](#keeper)'s methods to update the state. 

Handler functions return a result of type `sdk.Result`, which informs the application on whether the message was successfully processed:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/result.go#L15-L40

### Querier 

[`Queriers`](../building-modules/querier.md) are very similar to `handlers`, except they serve user queries to the state as opposed to processing transactions. A [query](../building-modules/messages-and-queries.md#queries) is initiated from an [interface](#interfaces) by an end-user who provides a `queryRoute` and some `data`. The query is then routed to the correct application's `querier` by `baseapp`'s `handleQueryCustom` method using `queryRoute`:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/baseapp/abci.go#L395-L453

The `Querier` of a module is defined in a file called `querier.go`, and consists of:

- A **switch function** `NewQuerier` to route the query to the appropriate `querier` function. This function returns a `querier` function, and is is registered in the [`AppModule`](#application-module-interface) to be used in the application's module manager to initialize the [application's query router](../core/baseapp.md#query-routing). See an example of such a switch from the [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice):
    +++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/internal/keeper/querier.go#L19-L32
- **One querier function for each data type defined by the module that needs to be queryable**. Developers write the query processing logic in these functions. This generally involves calling [`keeper`](#keeper)'s methods to query the state and marshalling it to JSON. 


### Keeper

[`Keepers`](../building-modules/keeper.md) are the gatekeepers of their module's store(s). To read or write in a module's store, it is mandatory to go through one of its `keeper`'s methods. This is ensured by the [object-capabilities](../core/ocap.md) model of the Cosmos SDK. Only objects that hold the key to a store can access it, and only the module's `keeper` should hold the key(s) to the module's store(s).

`Keepers` are generally defined in a file called `keeper.go`. It contains the `keeper`'s type definition and methods.

The `keeper` type definition generally consists of:

- **Key(s)** to the module's store(s) in the multistore.
- Reference to **other module's `keepers`**. Only needed if the `keeper` needs to access other module's store(s) (either to read or write from them).
- A reference to the application's **codec**. The `keeper` needs it to marshal structs before storing them, or to unmarshal them when it retrieves them, because stores only accept `[]bytes` as value.

Along with the type definition, the next important component of the `keeper.go` file is the `keeper`'s constructor function, `NewKeeper`. This function instantiates a new `keeper` of the type defined above, with a `codec`, store `keys` and potentially references to other modules' `keeper`s as parameters. The `NewKeeper` function is called from the [application's constructor](#constructor-function). The rest of the file defines the `keeper`'s methods, primarily getters and setters.

### Command-Line and REST Interfaces

Each module defines command-line commands and REST routes to be exposed to end-user via the [application's interfaces](#application-interfaces). This enables end-users to create messages of the types defined in the module, or to query the subset of the state managed by the module.

#### CLI

Generally, the [commands related to a module](../building-modules/module-interfaces.md#cli) are defined in a folder called `client/cli` in the module's folder. The CLI divides commands in two category, transactions and queries, defined in `client/cli/tx.go` and `client/cli/query.go` respectively. Both commands are built on top of the [Cobra Library](https://github.com/spf13/cobra):

- Transactions commands let users generate new transactions so that they can be included in a block and eventually update the state. One command should be created for each [message type](#message-types) defined in the module. The command calls the constructor of the message with the parameters provided by the end-user, and wraps it into a transaction. The SDK handles signing and the addition of other transaction metadata. 
- Queries let users query the subset of the state defined by the module. Query commands forward queries to the [application's query router](../core/baseapp.md#query-routing), which routes them to the appropriate [querier](#querier) the `queryRoute` parameter supplied. 

#### REST

The [module's REST interface](../building-modules/module-interfaces.md#rest) lets users generate transactions and query the state through REST calls to the application's [light client daemon](../core/node.md#lcd) (LCD). REST routes are defined in a file `client/rest/rest.go`, which is composed of:

- A `RegisterRoutes` function, which registers each route defined in the file. This function is called from the [main application's interface](#application-interfaces) for each module used within the application. The router used in the SDK is [Gorilla's mux](https://github.com/gorilla/mux).
- Custom request type definitions for each query or transaction creation function that needs to be exposed. These custom request types build on the base `request` type of the Cosmos SDK:
    +++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/rest/rest.go#L47-L60
- One handler function for each request that can be routed to the given module. These functions implement the core logic necessary to serve the request.

## Application Interface

[Interfaces](../interfaces/interfaces-intro.md) let end-users interact with full-node clients. This means querying data from the full-node or creating and sending new transactions to be relayed by the full-node and eventually included in a block.

The main interface is the [Command-Line Interface](../interfaces/cli.md). The CLI of an SDK application is built by aggregating [CLI commands](#cli) defined in each of the modules used by the application. The CLI of an application generally has the `-cli` suffix (e.g. `appcli`), and defined in a file called `cmd/appcli/main.go`. The file contains:

- **A `main()` function**, which is executed to build the `appcli` interface client. This function prepares each command and adds them to the `rootCmd` before building them. At the root of `appCli`, the function adds generic commands like `status`, `keys` and `config`, query commands, tx commands and `rest-server`.
- **Query commands** are added by calling the `queryCmd` function, also defined in `appcli/main.go`. This function returns a Cobra command that contains the query commands defined in each of the application's modules (passed as an array of `sdk.ModuleClients` from the `main()` function), as well as some other lower level query commands such as block or validator queries. Query command are called by using the command `appcli query [query]` of the CLI. 
- **Transaction commands** are added by calling the `txCmd` function. Similar to `queryCmd`, the function  returns a Cobra command that contains the tx commands defined in each of the application's modules, as well as lower level tx commands like transaction signing or broadcasting. Tx commands are called by using the command `appcli tx [tx]` of the CLI.
- **A `registerRoutes` function**, which is called from the `main()` function when initializing the [application's light-client daemon (LCD)](../core/node.md#lcd) (i.e. `rest-server`). `registerRoutes` calls the `RegisterRoutes` function of each of the application's module, thereby registering the routes of the module to the lcd's router. The LCD can be started by running the following command `appcli rest-server`. 

See an example of an application's main command-line file from the [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice)

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go

## Dependencies and Makefile

This section is optional, as developers are free to choose their dependency manager and project building method. That said, the current most used framework for versioning control is [`go.mod`](https://github.com/golang/go/wiki/Modules). It ensures each of the libraries used throughout the application are imported with the correct version. See an example from the [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice): 

+++ https://github.com/cosmos/sdk-tutorials/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/go.mod#L1-L18

For building the application, a [Makefile](https://en.wikipedia.org/wiki/Makefile) is generally used. The Makefile primarily ensures that the `go.mod` is run before building the two entrypoints to the application, [`appd`](#node-client) and [`appcli`](#application-interface). See an example of Makefile from the [nameservice tutorial]()

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/Makefile

## Next {hide}

Learn more about the [Lifecycle of a transaction](./tx-lifecycle.md) {hide}
