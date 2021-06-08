<!--
order: 1
-->

# Anatomy of an SDK Application

This document describes the core parts of a Cosmos SDK application. Throughout the document, a placeholder application named `app` will be used. {synopsis}

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
- **A reference to an [`appCodec`](../core/encoding.md).** The application's `appCodec` is used to serialize and deserialize data structures in order to store them, as stores can only persist `[]bytes`. The default codec is [Protocol Buffers](../core/encoding.md).
- **A reference to a [`legacyAmino`](../core/encoding.md) codec.** Some parts of the SDK have not been migrated to use the `appCodec` above, and are still hardcoded to use Amino. Other parts explicity use Amino for backwards compatibility. For these reasons, the application still holds a reference to the legacy Amino codec. Please note that the Amino codec will be removed from the SDK in the upcoming releases.
- **A reference to a [module manager](../building-modules/module-manager.md#manager)** and a [basic module manager](../building-modules/module-manager.md#basicmanager). The module manager is an object that contains a list of the application's module. It facilitates operations related to these modules, like registering their [`Msg` service](../core/baseapp.md#msg-services) and [gRPC `Query` service](../core/baseapp.md#grpc-query-services), or setting the order of execution between modules for various functions like [`InitChainer`](#initchainer), [`BeginBlocker` and `EndBlocker`](#beginblocker-and-endblocker).

See an example of application type definition from `simapp`, the SDK's own app used for demo and testing purposes:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/simapp/app.go#L145-L187

### Constructor Function

This function constructs a new application of the type defined in the section above. It must fulfill the `AppCreator` signature in order to be used in the [`start` command](../core/node.md#start-command) of the application's daemon command.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/server/types/app.go#L48-L50

Here are the main actions performed by this function:

- Instantiate a new [`codec`](../core/encoding.md) and initialize the `codec` of each of the application's module using the [basic manager](../building-modules/module-manager.md#basicmanager)
- Instantiate a new application with a reference to a `baseapp` instance, a codec and all the appropriate store keys.
- Instantiate all the [`keeper`s](#keeper) defined in the application's `type` using the `NewKeeper` function of each of the application's modules. Note that `keepers` must be instantiated in the correct order, as the `NewKeeper` of one module might require a reference to another module's `keeper`.
- Instantiate the application's [module manager](../building-modules/module-manager.md#manager) with the [`AppModule`](#application-module-interface) object of each of the application's modules.
- With the module manager, initialize the application's [`Msg` services](../core/baseapp.md#msg-services), [gRPC `Query` services](../core/baseapp.md#grpc-query-services), [legacy `Msg` routes](../core/baseapp.md#routing) and [legacy query routes](../core/baseapp.md#query-routing). When a transaction is relayed to the application by Tendermint via the ABCI, it is routed to the appropriate module's [`Msg` service](#msg-services) using the routes defined here. Likewise, when a gRPC query request is received by the application, it is routed to the appropriate module's [`gRPC query service`](#grpc-query-services) using the gRPC routes defined here. The SDK still supports legacy `Msg`s and legacy Tendermint queries, which are routed using respectively the legacy `Msg` routes and the legacy query routes.
- With the module manager, register the [application's modules' invariants](../building-modules/invariants.md). Invariants are variables (e.g. total supply of a token) that are evaluated at the end of each block. The process of checking invariants is done via a special module called the [`InvariantsRegistry`](../building-modules/invariants.md#invariant-registry). The value of the invariant should be equal to a predicted value defined in the module. Should the value be different than the predicted one, special logic defined in the invariant registry will be triggered (usually the chain is halted). This is useful to make sure no critical bug goes unnoticed and produces long-lasting effects that would be hard to fix.
- With the module manager, set the order of execution between the `InitGenesis`, `BegingBlocker` and `EndBlocker` functions of each of the [application's modules](#application-module-interface). Note that not all modules implement these functions.
- Set the remainer of application's parameters:
  - [`InitChainer`](#initchainer): used to initialize the application when it is first started.
  - [`BeginBlocker`, `EndBlocker`](#beginblocker-and-endlbocker): called at the beginning and the end of every block).
  - [`anteHandler`](../core/baseapp.md#antehandler): used to handle fees and signature verification.
- Mount the stores.
- Return the application.

Note that this function only creates an instance of the app, while the actual state is either carried over from the `~/.appd/data` folder if the node is restarted, or generated from the genesis file if the node is started for the first time.

See an example of application constructor from `simapp`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/simapp/app.go#L198-L441

### InitChainer

The `InitChainer` is a function that initializes the state of the application from a genesis file (i.e. token balances of genesis accounts). It is called when the application receives the `InitChain` message from the Tendermint engine, which happens when the node is started at `appBlockHeight == 0` (i.e. on genesis). The application must set the `InitChainer` in its [constructor](#constructor-function) via the [`SetInitChainer`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetInitChainer) method.

In general, the `InitChainer` is mostly composed of the [`InitGenesis`](../building-modules/genesis.md#initgenesis) function of each of the application's modules. This is done by calling the `InitGenesis` function of the module manager, which in turn will call the `InitGenesis` function of each of the modules it contains. Note that the order in which the modules' `InitGenesis` functions must be called has to be set in the module manager using the [module manager's](../building-modules/module-manager.md) `SetOrderInitGenesis` method. This is done in the [application's constructor](#application-constructor), and the `SetOrderInitGenesis` has to be called before the `SetInitChainer`.

See an example of an `InitChainer` from `simapp`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/simapp/app.go#L464-L471

### BeginBlocker and EndBlocker

The SDK offers developers the possibility to implement automatic execution of code as part of their application. This is implemented through two function called `BeginBlocker` and `EndBlocker`. They are called when the application receives respectively the `BeginBlock` and `EndBlock` messages from the Tendermint engine, which happens at the beginning and at the end of each block. The application must set the `BeginBlocker` and `EndBlocker` in its [constructor](#constructor-function) via the [`SetBeginBlocker`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetBeginBlocker) and [`SetEndBlocker`](https://godoc.org/github.com/cosmos/cosmos-sdk/baseapp#BaseApp.SetEndBlocker) methods.

In general, the `BeginBlocker` and `EndBlocker` functions are mostly composed of the [`BeginBlock` and `EndBlock`](../building-modules/beginblock-endblock.md) functions of each of the application's modules. This is done by calling the `BeginBlock` and `EndBlock` functions of the module manager, which in turn will call the `BeginBLock` and `EndBlock` functions of each of the modules it contains. Note that the order in which the modules' `BegingBlock` and `EndBlock` functions must be called has to be set in the module manager using the `SetOrderBeginBlock` and `SetOrderEndBlock` methods respectively. This is done via the [module manager](../building-modules/module-manager.md) in the [application's constructor](#application-constructor), and the `SetOrderBeginBlock` and `SetOrderEndBlock` methods have to be called before the `SetBeginBlocker` and `SetEndBlocker` functions.

As a sidenote, it is important to remember that application-specific blockchains are deterministic. Developers must be careful not to introduce non-determinism in `BeginBlocker` or `EndBlocker`, and must also be careful not to make them too computationally expensive, as [gas](./gas-fees.md) does not constrain the cost of `BeginBlocker` and `EndBlocker` execution.

See an example of `BeginBlocker` and `EndBlocker` functions from `simapp`

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/simapp/app.go#L454-L462

### Register Codec

The `EncodingConfig` structure is the last important part of the `app.go` file. The goal of this structure is to define the codecs that will be used throughout the app.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/simapp/params/encoding.go#L9-L16

Here are descriptions of what each of the four fields means:

- `InterfaceRegistry`: The `InterfaceRegistry` is used by the Protobuf codec to handle interfaces that are encoded and decoded (we also say "unpacked") using [`google.protobuf.Any`](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/any.proto). `Any` could be thought as a struct that contains a `type_url` (name of a concrete type implementing the interface) and a `value` (its encoded bytes). `InterfaceRegistry` provides a mechanism for registering interfaces and implementations that can be safely unpacked from `Any`. Each of the application's modules implements the `RegisterInterfaces` method that can be used to register the module's own interfaces and implementations.
  - You can read more about Any in [ADR-19](../architecture/adr-019-protobuf-state-encoding.md#usage-of-any-to-encode-interfaces).
  - To go more into details, the SDK uses an implementation of the Protobuf specification called [`gogoprotobuf`](https://github.com/gogo/protobuf). By default, the [gogo protobuf implementation of `Any`](https://godoc.org/github.com/gogo/protobuf/types) uses [global type registration](https://github.com/gogo/protobuf/blob/master/proto/properties.go#L540) to decode values packed in `Any` into concrete Go types. This introduces a vulnerability where any malicious module in the dependency tree could registry a type with the global protobuf registry and cause it to be loaded and unmarshaled by a transaction that referenced it in the `type_url` field. For more information, please refer to [ADR-019](../architecture/adr-019-protobuf-state-encoding.md).
- `Marshaler`: The `Marshaler` is the default codec used throughout the SDK. It is composed of a `BinaryMarshaler` used to encode and decode state, and a `JSONMarshaler` used to output data to the users (for example in the [CLI](#cli)). By default, the SDK uses Protobuf as `Marshaler`.
- `TxConfig`: `TxConfig` defines an interface a client can utilize to generate an application-defined concrete transaction type. Currently, the SDK handles two transaction types: `SIGN_MODE_DIRECT` (which uses Protobuf binary as over-the-wire encoding) and `SIGN_MODE_LEGACY_AMINO_JSON` (which depends on Amino). Read more about transactions [here](../core/transactions.md).
- `Amino`: Some legacy parts of the SDK still use Amino for backwards-compatibility. Each module exposes a `RegisterLegacyAmino` method to register the module's specific types within Amino. This `Amino` codec should not be used by app developers anymore, and will be removed in future releases.

The SDK exposes a `MakeTestEncodingConfig` function used to create a `EncodingConfig` for the app constructor (`NewApp`). It uses Protobuf as a default `Marshaler`, and passes it down to the app's `appCodec` field. It also instantiates a legacy `Amino` codec inside the app's `legacyAmino` field.
NOTE: this function is marked deprecated and should only be used to create an app or in tests. We are working on refactoring codec management in a post Stargate release.

See an example of a `MakeCodecs` from `simapp`:

+++ https://github.com/cosmos/cosmos-sdk/blob/590358652cc1cbc13872ea1659187e073ea38e75/simapp/encoding.go#L8-L19

## Modules

[Modules](../building-modules/intro.md) are the heart and soul of SDK applications. They can be considered as state-machines within the state-machine. When a transaction is relayed from the underlying Tendermint engine via the ABCI to the application, it is routed by [`baseapp`](../core/baseapp.md) to the appropriate module in order to be processed. This paradigm enables developers to easily build complex state-machines, as most of the modules they need often already exist. For developers, most of the work involved in building an SDK application revolves around building custom modules required by their application that do not exist yet, and integrating them with modules that do already exist into one coherent application. In the application directory, the standard practice is to store modules in the `x/` folder (not to be confused with the SDK's `x/` folder, which contains already-built modules).

### Application Module Interface

Modules must implement [interfaces](../building-modules/module-manager.md#application-module-interfaces) defined in the Cosmos SDK, [`AppModuleBasic`](../building-modules/module-manager.md#appmodulebasic) and [`AppModule`](../building-modules/module-manager.md#appmodule). The former implements basic non-dependant elements of the module, such as the `codec`, while the latter handles the bulk of the module methods (including methods that require references to other modules' `keeper`s). Both the `AppModule` and `AppModuleBasic` types are defined in a file called `./module.go`.

`AppModule` exposes a collection of useful methods on the module that facilitates the composition of modules into a coherent application. These methods are are called from the `module manager`(../building-modules/module-manager.md#manager), which manages the application's collection of modules.

### `Msg` Services

Each module defines two [Protobuf services](https://developers.google.com/protocol-buffers/docs/proto#services): one `Msg` service to handle messages, and one gRPC `Query` service to handle queries. If we consider the module as a state-machine, then a `Msg` is a state transition. A `Msg` service is a Protobuf service defining all possible `Msg`s a module exposes. Note that `Msg`s are bundled in [`transactions`](../core/transactions.md), and each transaction contains one or multiple `messages`.

When a valid block of transactions is received by the full-node, Tendermint relays each one to the application via [`DeliverTx`](https://tendermint.com/docs/app-dev/abci-spec.html#delivertx). Then, the application handles the transaction:

1. Upon receiving the transaction, the application first unmarshalls it from `[]bytes`.
2. Then, it verifies a few things about the transaction like [fee payment and signatures](#gas-fees.md#antehandler) before extracting the `Msg`(s) contained in the transaction.
3. `Msg`s are encoded as Protobuf [`Any`s](#register-codec) via the `sdk.ServiceMsg` struct. By analyzing each `Any`'s `type_url`, baseapp's `msgServiceRouter` routes the `Msg` to the corresponding module's `Msg` service.
4. If the message is successfully processed, the state is updated.

For a more detailed look at a transaction lifecycle, click [here](./tx-lifecycle.md).

Module developers create custom `Msg`s when they build their own module. The general practice is to define all `Msg`s in a Protobuf service called `service Msg {}`, and define each `Msg` as a Protobuf service method, using the `rpc` keyword. These definitions usually reside in a `tx.proto` file. For example, the `x/bank` module defines two `Msg`s to allows users to transfer tokens:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/proto/cosmos/bank/v1beta1/tx.proto#L10-L17

These two `Msg`s are processed by the `Msg` service of the `x/bank` module, which ultimately calls the `keeper` of the `x/auth` module in order to update the state.

Each module should also implement the `RegisterServices` method as part of the [`AppModule` interface](#application-module-interface). This method should call the `RegisterMsgServer` function provided by the generated Protobuf code.

### gRPC `Query` Services

gRPC `Query` services are introduced in the v0.40 Stargate release. They allow users to query the state using [gRPC](https://grpc.io). They are enabled by default, and can be configued under the `grpc.enable` and `grpc.address` fields inside [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml).

gRPC `Query` services are defined in the module's Protobuf definition files, specifically inside `query.proto`. The `query.proto` definition file exposes a single `Query` [Protobuf service](https://developers.google.com/protocol-buffers/docs/proto#services). Each gRPC query endpoint corresponds to a service method, starting with the `rpc` keyword, inside the `Query` service.

Protobuf generates a `QueryServer` interface for each module, containing all the service methods. A module's [`keeper`](#keeper) then needs to implement this `QueryServer` interface, by providing the concrete implementation of each service method. This concrete implementation is the handler of the corresponding gRPC query endpoint.

Finally, each module should also implement the `RegisterServices` method as part of the [`AppModule` interface](#application-module-interface). This method should call the `RegisterQueryServer` function provided by the generated Protobuf code.

### Keeper

[`Keepers`](../building-modules/keeper.md) are the gatekeepers of their module's store(s). To read or write in a module's store, it is mandatory to go through one of its `keeper`'s methods. This is ensured by the [object-capabilities](../core/ocap.md) model of the Cosmos SDK. Only objects that hold the key to a store can access it, and only the module's `keeper` should hold the key(s) to the module's store(s).

`Keepers` are generally defined in a file called `keeper.go`. It contains the `keeper`'s type definition and methods.

The `keeper` type definition generally consists of:

- **Key(s)** to the module's store(s) in the multistore.
- Reference to **other module's `keepers`**. Only needed if the `keeper` needs to access other module's store(s) (either to read or write from them).
- A reference to the application's **codec**. The `keeper` needs it to marshal structs before storing them, or to unmarshal them when it retrieves them, because stores only accept `[]bytes` as value.

Along with the type definition, the next important component of the `keeper.go` file is the `keeper`'s constructor function, `NewKeeper`. This function instantiates a new `keeper` of the type defined above, with a `codec`, store `keys` and potentially references to other modules' `keeper`s as parameters. The `NewKeeper` function is called from the [application's constructor](#constructor-function). The rest of the file defines the `keeper`'s methods, primarily getters and setters.

### Command-Line, gRPC Services and REST Interfaces

Each module defines command-line commands, gRPC services and REST routes to be exposed to end-user via the [application's interfaces](#application-interfaces). This enables end-users to create messages of the types defined in the module, or to query the subset of the state managed by the module.

#### CLI

Generally, the [commands related to a module](../building-modules/module-interfaces.md#cli) are defined in a folder called `client/cli` in the module's folder. The CLI divides commands in two category, transactions and queries, defined in `client/cli/tx.go` and `client/cli/query.go` respectively. Both commands are built on top of the [Cobra Library](https://github.com/spf13/cobra):

- Transactions commands let users generate new transactions so that they can be included in a block and eventually update the state. One command should be created for each [message type](#message-types) defined in the module. The command calls the constructor of the message with the parameters provided by the end-user, and wraps it into a transaction. The SDK handles signing and the addition of other transaction metadata.
- Queries let users query the subset of the state defined by the module. Query commands forward queries to the [application's query router](../core/baseapp.md#query-routing), which routes them to the appropriate [querier](#querier) the `queryRoute` parameter supplied.

#### gRPC

[gRPC](https://grpc.io) is a modern open source high performance RPC framework that has support in multiple languages. It is the recommended way for external clients (such as wallets, browsers and other backend services) to interact with a node.

Each module can expose gRPC endpoints, called [service methods](https://grpc.io/docs/what-is-grpc/core-concepts/#service-definition) and are defined in the [module's Protobuf `query.proto` file](#grpc-query-services). A service method is defined by its name, input arguments and output response. The module then needs to:

- define a `RegisterGRPCGatewayRoutes` method on `AppModuleBasic` to wire the client gRPC requests to the correct handler inside the module.
- for each service method, define a corresponding handler. The handler implements the core logic necessary to serve the gRPC request, and is located in the `keeper/grpc_query.go` file.

#### gRPC-gateway REST Endpoints

Some external clients may not wish to use gRPC. The SDK provides in this case a gRPC gateway service, which exposes each gRPC service as a correspoding REST endpoint. Please refer to the [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/) documentation to learn more.

The REST endpoints are defined in the Protobuf files, along with the gRPC services, using Protobuf annotations. Modules that want to expose REST queries should add `google.api.http` annotations to their `rpc` methods. By default, all REST endpoints defined in the SDK have an URL starting with the `/cosmos/` prefix.

The SDK also provides a development endpoint to generate [Swagger](https://swagger.io/) definition files for these REST endpoints. This endpoint can be enabled inside the [`app.toml`](../run-node/run-node.md#configuring-the-node-using-apptoml) config file, under the `api.swagger` key.

#### Legacy API REST Endpoints

The [module's Legacy REST interface](../building-modules/module-interfaces.md#legacy-rest) lets users generate transactions and query the state through REST calls to the application's Legacy API Service. REST routes are defined in a file `client/rest/rest.go`, which is composed of:

- A `RegisterRoutes` function, which registers each route defined in the file. This function is called from the [main application's interface](#application-interfaces) for each module used within the application. The router used in the SDK is [Gorilla's mux](https://github.com/gorilla/mux).
- Custom request type definitions for each query or transaction creation function that needs to be exposed. These custom request types build on the base `request` type of the Cosmos SDK:
  +++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/rest/rest.go#L62-L76
- One handler function for each request that can be routed to the given module. These functions implement the core logic necessary to serve the request.

These Legacy API endpoints are present in the SDK for backward compatibility purposes and will be removed in the next release.

## Application Interface

[Interfaces](#command-line-grpc-services-and-rest-interfaces) let end-users interact with full-node clients. This means querying data from the full-node or creating and sending new transactions to be relayed by the full-node and eventually included in a block.

The main interface is the [Command-Line Interface](../core/cli.md). The CLI of an SDK application is built by aggregating [CLI commands](#cli) defined in each of the modules used by the application. The CLI of an application is the same as the daemon (e.g. `appd`), and defined in a file called `appd/main.go`. The file contains:

- **A `main()` function**, which is executed to build the `appd` interface client. This function prepares each command and adds them to the `rootCmd` before building them. At the root of `appd`, the function adds generic commands like `status`, `keys` and `config`, query commands, tx commands and `rest-server`.
- **Query commands** are added by calling the `queryCmd` function. This function returns a Cobra command that contains the query commands defined in each of the application's modules (passed as an array of `sdk.ModuleClients` from the `main()` function), as well as some other lower level query commands such as block or validator queries. Query command are called by using the command `appd query [query]` of the CLI.
- **Transaction commands** are added by calling the `txCmd` function. Similar to `queryCmd`, the function returns a Cobra command that contains the tx commands defined in each of the application's modules, as well as lower level tx commands like transaction signing or broadcasting. Tx commands are called by using the command `appd tx [tx]` of the CLI.

See an example of an application's main command-line file from the [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice)

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/cmd/nscli/main.go

## Dependencies and Makefile

::: warning
A patch introduced in `go-grpc v1.34.0` made gRPC incompatible with the `gogoproto` library, making some [gRPC queries](https://github.com/cosmos/cosmos-sdk/issues/8426) panic. As such, the SDK requires that `go-grpc <=v1.33.2` is installed in your `go.mod`.

To make sure that gRPC is working properly, it is **highly recommended** to add the following line in your application's `go.mod`:

```
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2
```

Please see [issue #8392](https://github.com/cosmos/cosmos-sdk/issues/8392) for more info.
:::

This section is optional, as developers are free to choose their dependency manager and project building method. That said, the current most used framework for versioning control is [`go.mod`](https://github.com/golang/go/wiki/Modules). It ensures each of the libraries used throughout the application are imported with the correct version. See an example from the [nameservice tutorial](https://github.com/cosmos/sdk-tutorials/tree/master/nameservice):

+++ https://github.com/cosmos/sdk-tutorials/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/go.mod#L1-L18

For building the application, a [Makefile](https://en.wikipedia.org/wiki/Makefile) is generally used. The Makefile primarily ensures that the `go.mod` is run before building the two entrypoints to the application, [`appd`](#node-client) and [`appd`](#application-interface). See an example of Makefile from the [nameservice tutorial]()

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/Makefile

## Next {hide}

Learn more about the [Lifecycle of a transaction](./tx-lifecycle.md) {hide}
