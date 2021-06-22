<!--
order: 2
-->

# Module Manager

Cosmos SDK modules need to implement the [`AppModule` interfaces](#application-module-interfaces), in order to be managed by the application's [module manager](#module-manager). The module manager plays an important role in [`message` and `query` routing](../core/baseapp.md#routing), and allows application developers to set the order of execution of a variety of functions like [`BeginBlocker` and `EndBlocker`](../basics/app-anatomy.md#begingblocker-and-endblocker). {synopsis}

## Pre-requisite Readings

- [Introduction to SDK Modules](./intro.md) {prereq}

## Application Module Interfaces

Application module interfaces exist to facilitate the composition of modules together to form a functional SDK application. There are 3 main application module interfaces:

- [`AppModuleBasic`](#appmodulebasic) for independent module functionalities.
- [`AppModule`](#appmodule) for inter-dependent module functionalities (except genesis-related functionalities).
- [`AppModuleGenesis`](#appmodulegenesis) for inter-dependent genesis-related module functionalities.

The `AppModuleBasic` interface exists to define independent methods of the module, i.e. those that do not depend on other modules in the application. This allows for the construction of the basic application structure early in the application definition, generally in the `init()` function of the [main application file](../basics/app-anatomy.md#core-application-file).

The `AppModule` interface exists to define inter-dependent module methods. Many modules need to interract with other modules, typically through [`keeper`s](./keeper.md), which means there is a need for an interface where modules list their `keeper`s and other methods that require a reference to another module's object. `AppModule` interface also enables the module manager to set the order of execution between module's methods like `BeginBlock` and `EndBlock`, which is important in cases where the order of execution between modules matters in the context of the application.

Lastly the interface for genesis functionality `AppModuleGenesis` is separated out from full module functionality `AppModule` so that modules which
are only used for genesis can take advantage of the `Module` patterns without having to define many placeholder functions.

### `AppModuleBasic`

The `AppModuleBasic` interface defines the independent methods modules need to implement.

+++ https://github.com/cosmos/cosmos-sdk/blob/325be6ff215db457c6fc7668109640cd7fdac461/types/module/module.go#L49-L63

Let us go through the methods:

- `Name()`: Returns the name of the module as a `string`.
- `RegisterLegacyAminoCodec(*codec.LegacyAmino)`: Registers the `amino` codec for the module, which is used to marshal and unmarshal structs to/from `[]byte` in order to persist them in the module's `KVStore`.
- `RegisterInterfaces(codectypes.InterfaceRegistry)`: Registers a module's interface types and their concrete implementations as `proto.Message`.
- `DefaultGenesis(codec.JSONMarshaler)`: Returns a default [`GenesisState`](./genesis.md#genesisstate) for the module, marshalled to `json.RawMessage`. The default `GenesisState` need to be defined by the module developer and is primarily used for testing.
- `ValidateGenesis(codec.JSONMarshaler, client.TxEncodingConfig, json.RawMessage)`: Used to validate the `GenesisState` defined by a module, given in its `json.RawMessage` form. It will usually unmarshall the `json` before running a custom [`ValidateGenesis`](./genesis.md#validategenesis) function defined by the module developer.
- `RegisterRESTRoutes(client.Context, *mux.Router)`: Registers the REST routes for the module. These routes will be used to map REST request to the module in order to process them. See [gRPC and REST](../core/grpc_rest.md) for more.
- `RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)`: Registers gRPC routes for the module.
- `GetTxCmd()`: Returns the root [`Tx` command](./module-interfaces.md#tx) for the module. The subcommands of this root command are used by end-users to generate new transactions containing [`message`s](./messages-and-queries.md#queries) defined in the module.
- `GetQueryCmd()`: Return the root [`query` command](./module-interfaces.md#query) for the module. The subcommands of this root command are used by end-users to generate new queries to the subset of the state defined by the module.

All the `AppModuleBasic` of an application are managed by the [`BasicManager`](#basicmanager).

### `AppModuleGenesis`

The `AppModuleGenesis` interface is a simple embedding of the `AppModuleBasic` interface with two added methods.

+++ https://github.com/cosmos/cosmos-sdk/blob/325be6ff215db457c6fc7668109640cd7fdac461/types/module/module.go#L152-L158

Let us go through the two added methods:

- `InitGenesis(sdk.Context, codec.JSONMarshaler, json.RawMessage)`: Initializes the subset of the state managed by the module. It is called at genesis (i.e. when the chain is first started).
- `ExportGenesis(sdk.Context, codec.JSONMarshaler)`: Exports the latest subset of the state managed by the module to be used in a new genesis file. `ExportGenesis` is called for each module when a new chain is started from the state of an existing chain.

It does not have its own manager, and exists separately from [`AppModule`](#appmodule) only for modules that exist only to implement genesis functionalities, so that they can be managed without having to implement all of `AppModule`'s methods. If the module is not only used during genesis, `InitGenesis(sdk.Context, codec.JSONMarshaler, json.RawMessage)` and `ExportGenesis(sdk.Context,  codec.JSONMarshaler)` will generally be defined as methods of the concrete type implementing hte `AppModule` interface.

### `AppModule`

The `AppModule` interface defines the inter-dependent methods modules need to implement.

+++ https://github.com/cosmos/cosmos-sdk/blob/b4cce159bcc6a32ac78245c6866dd87c73f3720d/types/module/module.go#L160-L182

`AppModule`s are managed by the [module manager](#manager). This interface embeds the `AppModuleGenesis` interface so that the manager can access all the independent and genesis inter-dependent methods of the module. This means that a concrete type implementing the `AppModule` interface must either implement all the methods of `AppModuleGenesis` (and by extension `AppModuleBasic`), or include a concrete type that does as parameter.

Let us go through the methods of `AppModule`:

- `RegisterInvariants(sdk.InvariantRegistry)`: Registers the [`invariants`](./invariants.md) of the module. If the invariants deviates from its predicted value, the [`InvariantRegistry`](./invariants.md#registry) triggers appropriate logic (most often the chain will be halted).
- `Route()`: Returns the route for [`message`s](./messages-and-queries.md#messages) to be routed to the module by [`BaseApp`](../core/baseapp.md#message-routing).
- `QuerierRoute()` (deprecated): Returns the name of the module's query route, for [`queries`](./messages-and-queries.md#queries) to be routes to the module by [`BaseApp`](../core/baseapp.md#query-routing).
- `LegacyQuerierHandler(*codec.LegacyAmino)` (deprecated): Returns a [`querier`](./query-services.md#legacy-queriers) given the query `path`, in order to process the `query`.
- `RegisterServices(Configurator)`: Allows a module to register services.
- `BeginBlock(sdk.Context, abci.RequestBeginBlock)`: This method gives module developers the option to implement logic that is automatically triggered at the beginning of each block. Implement empty if no logic needs to be triggered at the beginning of each block for this module.
- `EndBlock(sdk.Context, abci.RequestEndBlock)`: This method gives module developers the option to implement logic that is automatically triggered at the end of each block. This is also where the module can inform the underlying consensus engine of validator set changes (e.g. the `staking` module). Implement empty if no logic needs to be triggered at the end of each block for this module.


### Implementing the Application Module Interfaces

Typically, the various application module interfaces are implemented in a file called `module.go`, located in the module's folder (e.g. `./x/module/module.go`).

Almost every module needs to implement the `AppModuleBasic` and `AppModule` interfaces. If the module is only used for genesis, it will implement `AppModuleGenesis` instead of `AppModule`. The concrete type that implements the interface can add parameters that are required for the implementation of the various methods of the interface. For example, the `Route()` function often calls a `NewHandler(k keeper)` function defined in [`handler.go`](./msg-services.md#handler-type) and therefore needs to pass the module's [`keeper`](./keeper.md) as a parameter.

```go
// example
type AppModule struct {
	AppModuleBasic
	keeper       Keeper
}
```

In the example above, you can see that the `AppModule` concrete type references an `AppModuleBasic`, and not an `AppModuleGenesis`. That is because `AppModuleGenesis` only needs to be implemented in modules that focus on genesis-related functionalities. In most modules, the concrete `AppModule` type will have a reference to an `AppModuleBasic` and implement the two added methods of `AppModuleGenesis` directly in the `AppModule` type.

If no parameter is required (which is often the case for `AppModuleBasic`), just declare an empty concrete type like so:

```go
type AppModuleBasic struct{}
```

## Module Managers

Module managers are used to manage collections of `AppModuleBasic` and `AppModule`.

### `BasicManager`

The `BasicManager` is a structure that lists all the `AppModuleBasic` of an application:

+++ https://github.com/cosmos/cosmos-sdk/blob/325be6ff215db457c6fc7668109640cd7fdac461/types/module/module.go#L65-L66

It implements the following methods:

- `NewBasicManager(modules ...AppModuleBasic)`: Constructor function. It takes a list of the application's `AppModuleBasic` and builds a new `BasicManager`. This function is generally called in the `init()` function of [`app.go`](../basics/app-anatomy.md#core-application-file) to quickly initialize the independent elements of the application's modules (click [here](https://github.com/cosmos/gaia/blob/master/app/app.go#L59-L74) to see an example).
- `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)`: Registers the [`codec.LegacyAmino`s](../core/encoding.md#amino) of each of the application's `AppModuleBasic`. This function is usually called early on in the [application's construction](../basics/app-anatomy.md#constructor).
- `RegisterInterfaces(registry codectypes.InterfaceRegistry)`: Registers interface types and implementations of each of the application's `AppModuleBasic`.
- `DefaultGenesis(cdc codec.JSONMarshaler)`: Provides default genesis information for modules in the application by calling the [`DefaultGenesis(cdc codec.JSONMarshaler)`](./genesis.md#defaultgenesis) function of each module. It is used to construct a default genesis file for the application.
- `ValidateGenesis(cdc codec.JSONMarshaler, txEncCfg client.TxEncodingConfig, genesis map[string]json.RawMessage)`: Validates the genesis information modules by calling the [`ValidateGenesis(codec.JSONMarshaler, client.TxEncodingConfig, json.RawMessage)`](./genesis.md#validategenesis) function of each module.
- `RegisterRESTRoutes(ctx client.Context, rtr *mux.Router)`: Registers REST routes for modules by calling the [`RegisterRESTRoutes`](./module-interfaces.md#register-routes) function of each module. This function is usually called function from the `main.go` function of the [application's command-line interface](../core/cli.md).
- `RegisterGRPCGatewayRoutes(clientCtx client.Context, rtr *runtime.ServeMux)`: Registers gRPC routes for modules.
- `AddTxCommands(rootTxCmd *cobra.Command)`: Adds modules' transaction commands to the application's [`rootTxCommand`](../core/cli.md#transaction-commands). This function is usually called function from the `main.go` function of the [application's command-line interface](../core/cli.md).
- `AddQueryCommands(rootQueryCmd *cobra.Command)`: Adds modules' query commands to the application's [`rootQueryCommand`](../core/cli.md#query-commands). This function is usually called function from the `main.go` function of the [application's command-line interface](../core/cli.md).

### `Manager`

The `Manager` is a structure that holds all the `AppModule` of an application, and defines the order of execution between several key components of these modules:

+++ https://github.com/cosmos/cosmos-sdk/blob/325be6ff215db457c6fc7668109640cd7fdac461/types/module/module.go#L223-L231

The module manager is used throughout the application whenever an action on a collection of modules is required. It implements the following methods:

- `NewManager(modules ...AppModule)`: Constructor function. It takes a list of the application's `AppModule`s and builds a new `Manager`. It is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderInitGenesis(moduleNames ...string)`: Sets the order in which the [`InitGenesis`](./genesis.md#initgenesis) function of each module will be called when the application is first started. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderExportGenesis(moduleNames ...string)`: Sets the order in which the [`ExportGenesis`](./genesis.md#exportgenesis) function of each module will be called in case of an export. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderBeginBlockers(moduleNames ...string)`: Sets the order in which the `BeginBlock()` function of each module will be called at the beginning of each block. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderEndBlockers(moduleNames ...string)`: Sets the order in which the `EndBlock()` function of each module will be called at the end of each block. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `RegisterInvariants(ir sdk.InvariantRegistry)`: Registers the [invariants](./invariants.md) of each module.
- `RegisterRoutes(router sdk.Router, queryRouter sdk.QueryRouter, legacyQuerierCdc *codec.LegacyAmino)`: Registers legacy [`Msg`](./messages-and-queries.md#messages) and [`querier`](./query-services.md#legacy-queriers) routes. 
- `RegisterServices(cfg Configurator)`: Registers all module services.
- `InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, genesisData map[string]json.RawMessage)`: Calls the [`InitGenesis`](./genesis.md#initgenesis) function of each module when the application is first started, in the order defined in `OrderInitGenesis`. Returns an `abci.ResponseInitChain` to the underlying consensus engine, which can contain validator updates.
- `ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler)`: Calls the [`ExportGenesis`](./genesis.md#exportgenesis) function of each module, in the order defined in `OrderExportGenesis`. The export constructs a genesis file from a previously existing state, and is mainly used when a hard-fork upgrade of the chain is required.
- `BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock)`: At the beginning of each block, this function is called from [`BaseApp`](../core/baseapp.md#beginblock) and, in turn, calls the [`BeginBlock`](./beginblock-endblock.md) function of each module, in the order defined in `OrderBeginBlockers`. It creates a child [context](../core/context.md) with an event manager to aggregate [events](../core/events.md) emitted from all modules. The function returns an `abci.ResponseBeginBlock` which contains the aforementioned events.
- `EndBlock(ctx sdk.Context, req abci.RequestEndBlock)`: At the end of each block, this function is called from [`BaseApp`](../core/baseapp.md#endblock) and, in turn, calls the [`EndBlock`](./beginblock-endblock.md) function of each module, in the order defined in `OrderEndBlockers`. It creates a child [context](../core/context.md) with an event manager to aggregate [events](../core/events.md) emitted from all modules. The function returns an `abci.ResponseEndBlock` which contains the aforementioned events, as well as validator set updates (if any).

Here's an example of a concrete integration within an application:

+++ https://github.com/cosmos/cosmos-sdk/blob/2323f1ac0e9a69a0da6b43693061036134193464/simapp/app.go#L315-L362

## Next {hide}

Learn more about [`message`s and `queries`](./messages-and-queries.md) {hide}
