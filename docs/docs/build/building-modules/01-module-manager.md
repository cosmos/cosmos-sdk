---
sidebar_position: 1
---

# Module Manager

:::note Synopsis
Cosmos SDK modules need to implement the [`AppModule` interfaces](#application-module-interfaces), in order to be managed by the application's [module manager](#module-manager). The module manager plays an important role in [`message` and `query` routing](../../learn/advanced/00-baseapp.md#routing), and allows application developers to set the order of execution of a variety of functions like [`PreBlocker`](../../learn/beginner/00-app-anatomy#preblocker) and [`BeginBlocker` and `EndBlocker`](../../learn/beginner/00-app-anatomy.md#begingblocker-and-endblocker).
:::

:::note Pre-requisite Readings

* [Introduction to Cosmos SDK Modules](./00-intro.md)

:::

## Application Module Interfaces

Application module interfaces exist to facilitate the composition of modules together to form a functional Cosmos SDK application.

:::note

It is recommended to implement interfaces from the [Core API](https://docs.cosmos.network/main/architecture/adr-063-core-module-api) `appmodule` package. This makes modules less dependent on the SDK.
For legacy reason modules can still implement interfaces from the SDK `module` package.
:::

There are 2 main application module interfaces:

* [`appmodule.AppModule` / `module.AppModule`](#appmodule) for inter-dependent module functionalities (except genesis-related functionalities).
* (legacy) [`module.AppModuleBasic`](#appmodulebasic) for independent module functionalities. New modules can use `module.CoreAppModuleBasicAdaptor` instead.

The above interfaces are mostly embedding smaller interfaces (extension interfaces), that defines specific functionalities:

* (legacy) `module.HasName`: Allows the module to provide its own name for legacy purposes.
* (legacy) [`module.HasGenesisBasics`](#modulehasgenesisbasics): The legacy interface for stateless genesis methods.
* [`module.HasGenesis`](#modulehasgenesis) for inter-dependent genesis-related module functionalities.
* [`module.HasABCIGenesis`](#modulehasabcigenesis) for inter-dependent genesis-related module functionalities.
* [`appmodule.HasGenesis` / `module.HasGenesis`](#appmodulehasgenesis): The extension interface for stateful genesis methods.
* [`appmodule.HasPreBlocker`](#haspreblocker): The extension interface that contains information about the `AppModule` and `PreBlock`.
* [`appmodule.HasBeginBlocker`](#hasbeginblocker): The extension interface that contains information about the `AppModule` and `BeginBlock`.
* [`appmodule.HasEndBlocker`](#hasendblocker): The extension interface that contains information about the `AppModule` and `EndBlock`.
* [`appmodule.HasPrecommit`](#hasprecommit): The extension interface that contains information about the `AppModule` and `Precommit`.
* [`appmodule.HasPrepareCheckState`](#haspreparecheckstate): The extension interface that contains information about the `AppModule` and `PrepareCheckState`.
* [`appmodule.HasService` / `module.HasServices`](#hasservices): The extension interface for modules to register services.
* [`module.HasABCIEndBlock`](#hasabciendblock): The extension interface that contains information about the `AppModule`, `EndBlock` and returns an updated validator set.
* (legacy) [`module.HasInvariants`](#hasinvariants): The extension interface for registering invariants.
* (legacy) [`module.HasConsensusVersion`](#hasconsensusversion): The extension interface for declaring a module consensus version.

The `AppModuleBasic` interface exists to define independent methods of the module, i.e. those that do not depend on other modules in the application. This allows for the construction of the basic application structure early in the application definition, generally in the `init()` function of the [main application file](../../learn/beginner/00-app-anatomy.md#core-application-file).

The `AppModule` interface exists to define inter-dependent module methods. Many modules need to interact with other modules, typically through [`keeper`s](./06-keeper.md), which means there is a need for an interface where modules list their `keeper`s and other methods that require a reference to another module's object. `AppModule` interface extension, such as `HasBeginBlocker` and `HasEndBlocker`, also enables the module manager to set the order of execution between module's methods like `BeginBlock` and `EndBlock`, which is important in cases where the order of execution between modules matters in the context of the application.

The usage of extension interfaces allows modules to define only the functionalities they need. For example, a module that does not need an `EndBlock` does not need to define the `HasEndBlocker` interface and thus the `EndBlock` method. `AppModule` and `AppModuleGenesis` are voluntarily small interfaces, that can take advantage of the `Module` patterns without having to define many placeholder functions.

### `AppModuleBasic`

:::note
Use `module.CoreAppModuleBasicAdaptor` instead for creating an `AppModuleBasic` from an `appmodule.AppModule`.
:::

The `AppModuleBasic` interface defines the independent methods modules need to implement.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L56-L61
```

* `RegisterLegacyAminoCodec(*codec.LegacyAmino)`: Registers the `amino` codec for the module, which is used to marshal and unmarshal structs to/from `[]byte` in order to persist them in the module's `KVStore`.
* `RegisterInterfaces(codectypes.InterfaceRegistry)`: Registers a module's interface types and their concrete implementations as `proto.Message`.
* `RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)`: Registers gRPC routes for the module.

All the `AppModuleBasic` of an application are managed by the [`BasicManager`](#basicmanager).

### `HasName`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L66-L68
```

* `HasName` is an interface that has a method `Name()`. This method returns the name of the module as a `string`.

### Genesis

:::tip
For easily creating an `AppModule` that only has genesis functionalities, use `module.GenesisOnlyAppModule`.
:::

#### `module.HasGenesisBasics`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L71-L74
```

Let us go through the methods:

* `DefaultGenesis(codec.JSONCodec)`: Returns a default [`GenesisState`](./08-genesis.md#genesisstate) for the module, marshalled to `json.RawMessage`. The default `GenesisState` need to be defined by the module developer and is primarily used for testing.
* `ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage)`: Used to validate the `GenesisState` defined by a module, given in its `json.RawMessage` form. It will usually unmarshall the `json` before running a custom [`ValidateGenesis`](./08-genesis.md#validategenesis) function defined by the module developer.

#### `module.HasGenesis`

`HasGenesis` is an extension interface for allowing modules to implement genesis functionalities.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/6ce2505/types/module/module.go#L184-L189
```

#### `module.HasABCIGenesis`

`HasABCIGenesis` is an extension interface for allowing modules to implement genesis functionalities and returns validator set updates.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/6ce2505/types/module/module.go#L191-L196
```

#### `appmodule.HasGenesis`

:::warning
`appmodule.HasGenesis` is experimental and should be considered unstable, it is recommended to not use this interface at this time.
:::

```go reference
https://github.com/cosmos/cosmos-sdk/blob/6ce2505/core/appmodule/genesis.go#L8-L25
```

### `AppModule`

The `AppModule` interface defines a module. Modules can declare their functionalities by implementing extensions interfaces.
`AppModule`s are managed by the [module manager](#manager), which checks which extension interfaces are implemented by the module.

#### `appmodule.AppModule`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/6afece6/core/appmodule/module.go#L11-L20
```

#### `module.AppModule`

:::note 
Previously the `module.AppModule` interface was containing all the methods that are defined in the extensions interfaces. This was leading to much boilerplate for modules that did not need all the functionalities.
:::

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L199-L206
```

### `HasInvariants`

This interface defines one method. It allows to checks if a module can register invariants.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L211-L214
```

* `RegisterInvariants(sdk.InvariantRegistry)`: Registers the [`invariants`](./07-invariants.md) of the module. If an invariant deviates from its predicted value, the [`InvariantRegistry`](./07-invariants.md#registry) triggers appropriate logic (most often the chain will be halted).

### `HasServices`

This interface defines one method. It allows to checks if a module can register invariants.

#### `appmodule.HasService`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/6afece6/core/appmodule/module.go#L22-L40
```

#### `module.HasServices`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L217-L220
```

* `RegisterServices(Configurator)`: Allows a module to register services.

### `HasConsensusVersion`

This interface defines one method for checking a module consensus version.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L223-L229
```

* `ConsensusVersion() uint64`: Returns the consensus version of the module.

### `HasPreBlocker`

The `HasPreBlocker` is an extension interface from `appmodule.AppModule`. All modules that have an `PreBlock` method implement this interface.

### `HasBeginBlocker`

The `HasBeginBlocker` is an extension interface from `appmodule.AppModule`. All modules that have an `BeginBlock` method implement this interface.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/core/appmodule/module.go#L73-L80
```

* `BeginBlock(context.Context) error`: This method gives module developers the option to implement logic that is automatically triggered at the beginning of each block.

### `HasEndBlocker`

The `HasEndBlocker` is an extension interface from `appmodule.AppModule`. All modules that have an `EndBlock` method implement this interface. If a module need to return validator set updates (staking), they can use `HasABCIEndBlock`

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/core/appmodule/module.go#L83-L89
```

* `EndBlock(context.Context) error`: This method gives module developers the option to implement logic that is automatically triggered at the end of each block.

### `HasABCIEndBlock`

The `HasABCIEndBlock` is an extension interface from `module.AppModule`. All modules that have an `EndBlock` which return validator set updates implement this interface.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L236-L239
```

* `EndBlock(context.Context) ([]abci.ValidatorUpdate, error)`: This method gives module developers the option to inform the underlying consensus engine of validator set changes (e.g. the `staking` module).

### `HasPrecommit`

`HasPrecommit` is an extension interface from `appmodule.AppModule`. All modules that have a `Precommit` method implement this interface.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/core/appmodule/module.go#L50-L53
```

* `Precommit(context.Context)`: This method gives module developers the option to implement logic that is automatically triggered during [`Commit'](../../learn/advanced/00-baseapp.md#commit) of each block using the [`finalizeblockstate`](../../learn/advanced/00-baseapp.md#state-updates) of the block to be committed. Implement empty if no logic needs to be triggered during `Commit` of each block for this module.

### `HasPrepareCheckState`

`HasPrepareCheckState` is an extension interface from `appmodule.AppModule`. All modules that have a `PrepareCheckState` method implement this interface.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/core/appmodule/module.go#L44-L47
```

* `PrepareCheckState(context.Context)`: This method gives module developers the option to implement logic that is automatically triggered during [`Commit'](../../learn/advanced/00-baseapp.md#commit) of each block using the [`checkState`](../../learn/advanced/00-baseapp.md#state-updates) of the next block. Implement empty if no logic needs to be triggered during `Commit` of each block for this module.

### Implementing the Application Module Interfaces

Typically, the various application module interfaces are implemented in a file called `module.go`, located in the module's folder (e.g. `./x/module/module.go`).

Almost every module needs to implement the `AppModuleBasic` and `AppModule` interfaces. If the module is only used for genesis, it will implement `AppModuleGenesis` instead of `AppModule`. The concrete type that implements the interface can add parameters that are required for the implementation of the various methods of the interface. For example, the `Route()` function often calls a `NewMsgServerImpl(k keeper)` function defined in `keeper/msg_server.go` and therefore needs to pass the module's [`keeper`](./06-keeper.md) as a parameter.

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

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L77
```

It implements the following methods:

* `NewBasicManager(modules ...AppModuleBasic)`: Constructor function. It takes a list of the application's `AppModuleBasic` and builds a new `BasicManager`. This function is generally called in the `init()` function of [`app.go`](../../learn/beginner/00-app-anatomy.md#core-application-file) to quickly initialize the independent elements of the application's modules (click [here](https://github.com/cosmos/gaia/blob/main/app/app.go#L59-L74) to see an example).
* `NewBasicManagerFromManager(manager *Manager, customModuleBasics map[string]AppModuleBasic)`: Contructor function. It creates a new `BasicManager` from a `Manager`. The `BasicManager` will contain all `AppModuleBasic` from the `AppModule` manager using `CoreAppModuleBasicAdaptor` whenever possible. Module's `AppModuleBasic` can be overridden by passing a custom AppModuleBasic map
* `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)`: Registers the [`codec.LegacyAmino`s](../../learn/advanced/05-encoding.md#amino) of each of the application's `AppModuleBasic`. This function is usually called early on in the [application's construction](../../learn/beginner/00-app-anatomy.md#constructor).
* `RegisterInterfaces(registry codectypes.InterfaceRegistry)`: Registers interface types and implementations of each of the application's `AppModuleBasic`.
* `DefaultGenesis(cdc codec.JSONCodec)`: Provides default genesis information for modules in the application by calling the [`DefaultGenesis(cdc codec.JSONCodec)`](./08-genesis.md#defaultgenesis) function of each module. It only calls the modules that implements the `HasGenesisBasics` interfaces.
* `ValidateGenesis(cdc codec.JSONCodec, txEncCfg client.TxEncodingConfig, genesis map[string]json.RawMessage)`: Validates the genesis information modules by calling the [`ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage)`](./08-genesis.md#validategenesis) function of modules implementing the `HasGenesisBasics` interface.
* `RegisterGRPCGatewayRoutes(clientCtx client.Context, rtr *runtime.ServeMux)`: Registers gRPC routes for modules.
* `AddTxCommands(rootTxCmd *cobra.Command)`: Adds modules' transaction commands (defined as `GetTxCmd() *cobra.Command`) to the application's [`rootTxCommand`](../../learn/advanced/07-cli.md#transaction-commands). This function is usually called function from the `main.go` function of the [application's command-line interface](../../learn/advanced/07-cli.md).
* `AddQueryCommands(rootQueryCmd *cobra.Command)`: Adds modules' query commands (defined as `GetQueryCmd() *cobra.Command`) to the application's [`rootQueryCommand`](../../learn/advanced/07-cli.md#query-commands). This function is usually called function from the `main.go` function of the [application's command-line interface](../../learn/advanced/07-cli.md).

### `Manager`

The `Manager` is a structure that holds all the `AppModule` of an application, and defines the order of execution between several key components of these modules:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/types/module/module.go#L278-L288
```

The module manager is used throughout the application whenever an action on a collection of modules is required. It implements the following methods:

* `NewManager(modules ...AppModule)`: Constructor function. It takes a list of the application's `AppModule`s and builds a new `Manager`. It is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderInitGenesis(moduleNames ...string)`: Sets the order in which the [`InitGenesis`](./08-genesis.md#initgenesis) function of each module will be called when the application is first started. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
  To initialize modules successfully, module dependencies should be considered. For example, the `genutil` module must occur after `staking` module so that the pools are properly initialized with tokens from genesis accounts, the `genutils` module must also occur after `auth` so that it can access the params from auth, IBC's `capability` module should be initialized before all other modules so that it can initialize any capabilities.
* `SetOrderExportGenesis(moduleNames ...string)`: Sets the order in which the [`ExportGenesis`](./08-genesis.md#exportgenesis) function of each module will be called in case of an export. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderPreBlockers(moduleNames ...string)`: Sets the order in which the `PreBlock()` function of each module will be called before `BeginBlock()` of all modules. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderBeginBlockers(moduleNames ...string)`: Sets the order in which the `BeginBlock()` function of each module will be called at the beginning of each block. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderEndBlockers(moduleNames ...string)`: Sets the order in which the `EndBlock()` function of each module will be called at the end of each block. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderPrecommiters(moduleNames ...string)`: Sets the order in which the `Precommit()` function of each module will be called during commit of each block. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderPrepareCheckStaters(moduleNames ...string)`: Sets the order in which the `PrepareCheckState()` function of each module will be called during commit of each block. This function is generally called from the application's main [constructor function](../../learn/beginner/00-app-anatomy.md#constructor-function).
* `SetOrderMigrations(moduleNames ...string)`: Sets the order of migrations to be run. If not set then migrations will be run with an order defined in `DefaultMigrationsOrder`.
* `RegisterInvariants(ir sdk.InvariantRegistry)`: Registers the [invariants](./07-invariants.md) of module implementing the `HasInvariants` interface.
* `RegisterServices(cfg Configurator)`: Registers the services of modules implementing the `HasServices` interface.
* `InitGenesis(ctx context.Context, cdc codec.JSONCodec, genesisData map[string]json.RawMessage)`: Calls the [`InitGenesis`](./08-genesis.md#initgenesis) function of each module when the application is first started, in the order defined in `OrderInitGenesis`. Returns an `abci.ResponseInitChain` to the underlying consensus engine, which can contain validator updates.
* `ExportGenesis(ctx context.Context, cdc codec.JSONCodec)`: Calls the [`ExportGenesis`](./08-genesis.md#exportgenesis) function of each module, in the order defined in `OrderExportGenesis`. The export constructs a genesis file from a previously existing state, and is mainly used when a hard-fork upgrade of the chain is required.
* `ExportGenesisForModules(ctx context.Context, cdc codec.JSONCodec, modulesToExport []string)`: Behaves the same as `ExportGenesis`, except takes a list of modules to export.
* `BeginBlock(ctx context.Context) error`: At the beginning of each block, this function is called from [`BaseApp`](../../learn/advanced/00-baseapp.md#beginblock) and, in turn, calls the [`BeginBlock`](./06-beginblock-endblock.md) function of each modules implementing the `appmodule.HasBeginBlocker` interface, in the order defined in `OrderBeginBlockers`. It creates a child [context](../../learn/advanced/02-context.md) with an event manager to aggregate [events](../../learn/advanced/08-events.md) emitted from each modules.
* `EndBlock(ctx context.Context) error`: At the end of each block, this function is called from [`BaseApp`](../../learn/advanced/00-baseapp.md#endblock) and, in turn, calls the [`EndBlock`](./06-beginblock-endblock.md) function of each modules implementing the `appmodule.HasEndBlocker` interface, in the order defined in `OrderEndBlockers`. It creates a child [context](../../learn/advanced/02-context.md) with an event manager to aggregate [events](../../learn/advanced/08-events.md) emitted from all modules. The function returns an `abci` which contains the aforementioned events, as well as validator set updates (if any).
* `EndBlock(context.Context) ([]abci.ValidatorUpdate, error)`: At the end of each block, this function is called from [`BaseApp`](../../learn/advanced/00-baseapp.md#endblock) and, in turn, calls the [`EndBlock`](./06-beginblock-endblock.md) function of each modules implementing the `module.HasABCIEndBlock` interface, in the order defined in `OrderEndBlockers`. It creates a child [context](../../learn/advanced/02-context.md) with an event manager to aggregate [events](../../learn/advanced/08-events.md) emitted from all modules. The function returns an `abci` which contains the aforementioned events, as well as validator set updates (if any).
* `Precommit(ctx context.Context)`: During [`Commit`](../../learn/advanced/00-baseapp.md#commit), this function is called from `BaseApp` immediately before the [`deliverState`](../../learn/advanced/00-baseapp.md#state-updates) is written to the underlying [`rootMultiStore`](../../learn/advanced/04-store.md#commitmultistore) and, in turn calls the `Precommit` function of each modules implementing the `HasPrecommit` interface, in the order defined in `OrderPrecommiters`. It creates a child [context](../../learn/advanced/02-context.md) where the underlying `CacheMultiStore` is that of the newly committed block's [`finalizeblockstate`](../../learn/advanced/00-baseapp.md#state-updates).
* `PrepareCheckState(ctx context.Context)`: During [`Commit`](../../learn/advanced/00-baseapp.md#commit), this function is called from `BaseApp` immediately after the [`deliverState`](../../learn/advanced/00-baseapp.md#state-updates) is written to the underlying [`rootMultiStore`](../../learn/advanced/04-store.md#commitmultistore) and, in turn calls the `PrepareCheckState` function of each module implementing the `HasPrepareCheckState` interface, in the order defined in `OrderPrepareCheckStaters`. It creates a child [context](../../learn/advanced/02-context.md) where the underlying `CacheMultiStore` is that of the next block's [`checkState`](../../learn/advanced/00-baseapp.md#state-updates). Writes to this state will be present in the [`checkState`](../../learn/advanced/00-baseapp.md#state-updates) of the next block, and therefore this method can be used to prepare the `checkState` for the next block.

Here's an example of a concrete integration within an `simapp`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/simapp/app.go#L510-L533
```

This is the same example from `runtime` (the package that powers app di):

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/runtime/module.go#L63
```

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/runtime/module.go#L85
```
