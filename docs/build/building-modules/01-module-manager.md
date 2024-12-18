---
sidebar_position: 1
---

# Module Manager

:::note Synopsis
Cosmos SDK modules need to implement the [`AppModule` interfaces](#application-module-interfaces), in order to be managed by the application's [module manager](#module-manager). The module manager plays an important role in [`message` and `query` routing](../../learn/advanced/00-baseapp.md#service-routers), and allows application developers to set the order of execution of a variety of functions like [`PreBlocker`, `BeginBlocker` and `EndBlocker`](https://docs.cosmos.network/main/learn/beginner/app-anatomy).
:::

:::note Pre-requisite Readings

* [Introduction to Cosmos SDK Modules](./00-intro.md)

:::

## Application Module Interfaces

Application module interfaces exist to facilitate the composition of modules together to form a functional Cosmos SDK application.

Those interface are defined in the `cosmossdk.io/core/appmodule` and `cosmossdk.io/core/appmodule/v2` packages.

:::note
The difference between appmodule and appmodule v2 is mainly the introduction of handlers from Cosmos SDK (server) v2. The rest of the API remains the same, and are simply aliases between the two packages.
:::

Following a list of all interfaces a module can implement:

* [`appmodule.AppModule`](#appmodule) is the main interface that defines a module. By default, a module does nothing. To add functionalities, a module can implement extension interfaces.
* [`appmodule.HasPreBlocker`](#haspreblocker): The extension interface that contains information about the `AppModule` and `PreBlock`.
* [`appmodule.HasBeginBlocker`](#hasbeginblocker): The extension interface that contains information about the `AppModule` and `BeginBlock`.
* [`appmodule.HasEndBlocker`](#hasendblocker): The extension interface that contains information about the `AppModule` and `EndBlock`.
* [`module.HasABCIEndBlock`](#hasendblocker): The extension interface that contains information about the `AppModule`, `EndBlock` and returns an updated validator set (Usually only needed by staking).
* [`appmodule.HasRegisterInterfaces`](#hasregisterinterfaces): The extension interface for modules to register their message types.
* [`appmodule.HasService`](#hasservices): The extension interface for modules to register services. Note, this interface is not exposed in core to avoid a gRPC dependency. However it is usable in an application.
* [`appmodule.HasAminoCodec`](#hasaminocodec): The extension interface for modules to support JSON encoding and decoding via `amino`.
* [`appmodule.HasMigrations`](#hasmigrations): The extension interface for registering module migrations.
    * [`appmodule.HasConsensusVersion`](#hasconsensusversion): The extension interface for declaring a module consensus version. It is usually not used alone, but in conjunction with `HasMigrations`.
* [`appmodule.HasGenesis`](#hasgenesis) for inter-dependent genesis-related module functionalities.
* [`appmodule.HasABCIGenesis`](#hasabcigenesis) for inter-dependent genesis-related module functionalities, with validator set updates (Usually only needed by staking).

The `AppModule` interface exists to define inter-dependent module methods. Many modules need to interact with other modules, typically through [`keeper`s](./06-keeper.md), which means there is a need for an interface where modules list their `keeper`s and other methods that require a reference to another module's object. `AppModule` interface extension, such as `HasBeginBlocker` and `HasEndBlocker`, also enables the module manager to set the order of execution between module's methods like `BeginBlock` and `EndBlock`, which is important in cases where the order of execution between modules matters in the context of the application.

The usage of extension interfaces allows modules to define only the functionalities they need. For example, a module that does not need an `EndBlock` does not need to define the `HasEndBlocker` interface and thus the `EndBlock` method. `AppModule` and `AppModuleGenesis` are voluntarily small interfaces, that can take advantage of the `Module` patterns without having to define many placeholder functions.

:::note legacy
Prior to the introduction of the `cosmossdk.io/core` package the interfaces were defined in the `types/module` package of the Cosmos SDK. Not all interfaces have been migrated to core. Those legacy interfaces are still supported for backward compatibility, but aren't described in this document and should not be used in new modules.
:::


### `AppModule`

The `AppModule` interface defines a module. Modules can declare their functionalities by implementing extensions interfaces.
`AppModule`s are managed by the [module manager](#manager), which checks which extension interfaces are implemented by the module.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/module.go#L10-L20
```

### `HasPreBlocker`

The `HasPreBlocker` is an extension interface from `appmodule.AppModule`. All modules that have an `PreBlock` method implement this interface.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/module.go#L22-L28
```

### `HasBeginBlocker`

The `HasBeginBlocker` is an extension interface from `appmodule.AppModule`. All modules that have an `BeginBlock` method implement this interface.
It gives module developers the option to implement logic that is automatically triggered at the beginning of each block.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/module.go#L30-L38
```

### `HasEndBlocker`

The `HasEndBlocker` is an extension interface from `appmodule.AppModule`. All modules that have an `EndBlock` method implement this interface. It gives module developers the option to implement logic that is automatically triggered at the end of each block.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/module.go#L40-L48
```

If a module needs to return validator set updates (staking), they can use `HasABCIEndBlock` (in v1).

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/types/module/module.go#L115-L119
```

Or, alternatively, `HasUpdateValidators` in v2:


```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/module.go#L87-L94
```

### `HasRegisterInterfaces`

The `HasRegisterInterfaces` is an extension interface from `appmodule.AppModule`. All modules that have a `RegisterInterfaces` method implement this interface. It allows modules to register their message types with their concrete implementations as `proto.Message`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/module.go#L103-L106
```

### `HasServices`

This interface defines one method. It allows to register and let module expose gRPC services.
This interface is not part of the `core` package to avoid a gRPC dependency, but is recognized by the module manager and [runtime](../building-apps/00-runtime.md).

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/module.go#L34-L53
```

### `HasAminoCodec`

The `HasAminoCodec` allows to register the `amino` codec for the module, which is used to marshal and unmarshal structs to/from `[]byte` in order to persist them in the module's `KVStore`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/module.go#L68-L72
```

### `module.HasGRPCGateway`

This interface is not part of the `core` package to avoid a gRPC dependency. It is used to register gRPC routes gateway routes for the module. [In v2, this will be done differently, and totally abstracted from modules and module manager](https://github.com/cosmos/cosmos-sdk/issues/22715)

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/types/module/module.go#L74-L77
```

### `HasMigrations`

The `HasMigrations` interface is used to register module migrations. Learn more about [module migrations](./13-upgrade.md).

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/migrations.go#L14-L21
```

### `HasConsensusVersion`

This interface defines one method for checking a module consensus version. It is mainly used in conjunction with `HasMigrations`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/migrations.go#L5-L12
```

### Genesis

#### `HasGenesis`

`HasGenesis` is an extension interface for allowing modules to implement genesis functionalities.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/genesis.go#L8-L19
```

Let us go through some of the methods:

* `DefaultGenesis()`: Returns a default [`GenesisState`](./08-genesis.md#genesisstate) for the module, marshalled to `json.RawMessage`. The default `GenesisState` need to be defined by the module developer and is primarily used for testing.
* `ValidateGenesis(data json.RawMessage) error`: Used to validate the `GenesisState` defined by a module, given in its `json.RawMessage` form. It will usually unmarshall the `json` before running a custom [`ValidateGenesis`](./08-genesis.md#validategenesis) function defined by the module developer.

In the same vein than `HasABCIEndBlock`, `HasABCIGenesis` is used to return validator set updates.

#### `HasABCIGenesis`

`HasABCIGenesis` is an extension interface for allowing modules to implement genesis functionalities and returns validator set updates.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/core/v1.0.0-alpha.6/core/appmodule/v2/genesis.go#L21-L31
```

### Implementing the Application Module Interfaces

Typically, the various application module interfaces are implemented in a file called `module.go`, located in the module's folder (e.g. `./x/module/module.go`).

Every module must implement the `AppModule` interface. If the module is only used for genesis, it will implement `HasGenesis` in addition of `AppModule`. The concrete type that implements the interface can add parameters that are required for the implementation of the various methods of the interface. 

```go
// example
type AppModule struct {
	keeper       Keeper
}

func (AppModule) IsAppModule() {}
func (AppModule) IsOnePerModuleType() {}
```

## Module Manager

The module manager is used to manage collections of `AppModule` and all the extensions interfaces.

### `Manager`

The `Manager` is a structure that holds all the `AppModule` of an application, and defines the order of execution between several key components of these modules:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/types/module/module.go#L121-L133
```

:::tip
Thanks to `runtime`, a user does not need to interact directly with the `Manager`. The `Manager` is used internally by the `runtime` to manage the modules of the application.
:::

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
* `RegisterServices(cfg Configurator)`: Registers the services of modules implementing the `HasServices` interface.
* `InitGenesis(ctx context.Context, genesisData map[string]json.RawMessage)`: Calls the [`InitGenesis`](./08-genesis.md#initgenesis) function of each module when the application is first started, in the order defined in `OrderInitGenesis`. Returns an `abci.InitChainResponse` to the underlying consensus engine, which can contain validator updates.
* `ExportGenesis(ctx context.Context)`: Calls the [`ExportGenesis`](./08-genesis.md#exportgenesis) function of each module, in the order defined in `OrderExportGenesis`. The export constructs a genesis file from a previously existing state, and is mainly used when a hard-fork upgrade of the chain is required.
* `ExportGenesisForModules(ctx context.Context, modulesToExport []string)`: Behaves the same as `ExportGenesis`, except takes a list of modules to export.
* `BeginBlock(ctx context.Context) error`: At the beginning of each block, this function is called from [`BaseApp`](../../learn/advanced/00-baseapp.md#beginblock) and, in turn, calls the [`BeginBlock`](./06-preblock-beginblock-endblock.md) function of each modules implementing the `appmodule.HasBeginBlocker` interface, in the order defined in `OrderBeginBlockers`.
* `EndBlock(ctx context.Context) error`: At the end of each block, this function is called from [`BaseApp`](../../learn/advanced/00-baseapp.md#endblock) and, in turn, calls the [`EndBlock`](./06-preblock-beginblock-endblock.md) function of each modules implementing the `appmodule.HasEndBlocker` interface, in the order defined in `OrderEndBlockers`.
* `EndBlock(context.Context) ([]abci.ValidatorUpdate, error)`: At the end of each block, this function is called from [`BaseApp`](../../learn/advanced/00-baseapp.md#endblock) and, in turn, calls the [`EndBlock`](./06-preblock-beginblock-endblock.md) function of each modules implementing the `appmodule.HasABCIEndBlock` interface, in the order defined in `OrderEndBlockers`. Extended implementation for modules that need to update the validator set (typically used by the staking module).
* (Optional) `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)`: Registers the [`codec.LegacyAmino`s](../../learn/advanced/05-encoding.md#amino) of each of the application module. This function is usually called early on in the [application's construction](../../learn/beginner/00-app-anatomy.md#constructor).
* `RegisterInterfaces(registry codectypes.InterfaceRegistry)`: Registers interface types and implementations of each of the application's `AppModule`.
* (Optional) `RegisterGRPCGatewayRoutes(clientCtx client.Context, rtr *runtime.ServeMux)`: Registers gRPC routes for modules.
* `DefaultGenesis(cdc codec.JSONCodec)`: Provides default genesis information for modules in the application by calling the [`DefaultGenesis(cdc codec.JSONCodec)`](./08-genesis.md#defaultgenesis) function of each module. It only calls the modules that implements the `HasGenesisBasics` interfaces.
* `ValidateGenesis(cdc codec.JSONCodec, txEncCfg client.TxEncodingConfig, genesis map[string]json.RawMessage)`: Validates the genesis information modules by calling the [`ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage)`](./08-genesis.md#validategenesis) function of modules implementing the `HasGenesisBasics` interface.

Here's an example of a concrete integration within [`runtime`](../building-apps/00-runtime.md)

```go reference
https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/runtime/module.go#L242-L244
```
