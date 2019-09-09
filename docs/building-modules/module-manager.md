# Module Manager and `AppModule` Interface

## Pre-requisite Reading

- [Introduction to SDK Modules](./intro.md)

## Synopsis

Cosmos SDK modules need to implement the [`AppModule` interfaces](#application-module-interfaces), in order to be managed by the application's [module manager](#module-manager). The module manager plays an important role in [`message` and `query` routing](../core/baseapp.md#routing), and allows the application developer to set the order of execution of a variety of functions like [`BeginBlocker` and `EndBlocker`](../basics/app-anatomy.md#begingblocker-and-endblocker). 

- [Application Module Interfaces](#application-module-interfaces)
	+ [`AppModuleBasic`](#appmodulebasic)
	+ [`AppModuleGenesis`](#appmodulegenesis)
	+ [`AppModule`](#appmodule)
	+ [Implementing the Application Module Interfaces](#implementing-the-application-module-interfaces)
- [Module Managers](#module-managers)
	+ [`BasicManager`](#basicmanager)
	+ [`Manager`](#manager)

## Application Module Interfaces

[Application module interfaces](https://github.com/cosmos/cosmos-sdk/blob/master/types/module/module.go) exist to facilitate the composition of modules together to form a functional SDK application. There are 3 main application module interfaces: 

- [`AppModuleBasic`](#appmodulebasic) for independent module functionalities.
- [`AppModule`](#appmodule) for inter-dependent module functionalities (except genesis-related functionalities).
- [`AppModuleGenesis`](#appmodulegenesis) for inter-dependent genesis-related module functionalities.

The `AppModuleBasic` interface exists to define independent methods of the module, i.e. those that do not depend on other modules in the application. This allows for the construction of the basic application structure early in the application definition, generally in the `init()` function of the [main application file](../basics/app-antomy.md#core-application-file).

The `AppModule` interface exists to define inter-dependent module methods. Many modules need to interract with other modules, typically through [`keeper`s](./keeper.md), which means there is a need for an interface where modules list their `keeper`s and other methods that require a reference to another module's object. `AppModule` interface also enables the module manager to set the order of execution between module's methods like `BeginBlock` and `EndBlock`, which is important in cases where the order of execution between modules matters in the context of the application. 

Lastly the interface for genesis functionality `AppModuleGenesis` is separated out from full module functionality `AppModule` so that modules which
are only used for genesis can take advantage of the `Module` patterns without having to define many placeholder functions. 

### `AppModuleBasic`

The [`AppModuleBasic`](https://github.com/cosmos/cosmos-sdk/blob/master/types/module/module.go#L45-L57) interface defines the independent methods modules need to implement. 

```go
type AppModuleBasic interface {
	Name() string
	RegisterCodec(*codec.Codec)

	// genesis
	DefaultGenesis() json.RawMessage
	ValidateGenesis(json.RawMessage) error

	// client functionality
	RegisterRESTRoutes(context.CLIContext, *mux.Router)
	GetTxCmd(*codec.Codec) *cobra.Command
	GetQueryCmd(*codec.Codec) *cobra.Command
}
```

Let us go through the methods:

- `Name()`: Returns the name of the module as a `string`.
- `RegisterCodec(*codec.Codec)`: Registers the `codec` for the module, which is used to marhsal and unmarshal structs to/from `[]byte` in order to persist them in the moduel's `KVStore`.
- `DefaultGenesis()`: Returns a default [`GenesisState`](./genesis.md#genesisstate) for the module, marshalled to `json.RawMessage`. The default `GenesisState` need to be defined by the module developer and is primarily used for testing. 
- `ValidateGenesis(json.RawMessage)`: Used to validate the `GenesisState` defined by a module, given in its `json.RawMessage` form. It will usually unmarshall the `json` before running a custom [`ValidateGenesis`](./genesis.md#validategenesis) function defined by the module developer. 
- `RegisterRESTRoutes(context.CLIContext, *mux.Router)`: Registers the REST routes for the module. These routes will be used to map REST request to the module in order to process them. See [../interfaces/rest.md] for more.
- `GetTxCmd(*codec.Codec)`: Returns the root [`Tx` command](./module-interfaces.md#tx) for the module. The subcommands of this root command are used by end-users to generate new transactions containing [`message`s](./messages-and-queries.md#queries) defined in the module. 
- `GetQueryCmd(*codec.Codec)`: Return the root [`query` command](./module-intefaces.md#query) for the module. The subcommands of this root command are used by end-users to generate new queries to the subset of the state defined by the module. 

All the `AppModuleBasic` of an application are managed by the [`BasicManager`](#basicmanager). 

### `AppModuleGenesis`

The [`AppModuleGenesis`](https://github.com/cosmos/cosmos-sdk/blob/master/types/module/module.go#L123-L127) interface is a simple embedding of the `AppModuleBasic` interface with two added methods.

```go
type AppModuleGenesis interface {
	AppModuleBasic
	InitGenesis(sdk.Context, json.RawMessage) []abci.ValidatorUpdate
	ExportGenesis(sdk.Context) json.RawMessage
}
```

Let us go through the two added methods:

- `InitGenesis(sdk.Context, json.RawMessage)`: Initializes the subset of the state managed by the module. It is called at genesis (i.e. when the chain is first started).
- `ExportGenesis(sdk.Context)`: Exports the latest subset of the state managed by the module to be used in a new genesis file. `ExportGenesis` is called for each module when a new chain is started from the state of an existing chain. 

It does not have its own manager, and exists separately from [`AppModule`](#appmodule) only for modules that exist only to implement genesis functionalities, so that they can be managed without having to implement all of `AppModule`'s methods. If the module is not only used during genesis, `InitGenesis(sdk.Context, json.RawMessage)` and `ExportGenesis(sdk.Context)` will generally be defined as methods of the concrete type implementing hte `AppModule` interface. 

### `AppModule`

The [`AppModule`](https://github.com/cosmos/cosmos-sdk/blob/master/types/module/module.go#L130-L144) interface defines the inter-dependent methods modules need to implement. 

```go
type AppModule interface {
	AppModuleGenesis

	// registers
	RegisterInvariants(sdk.InvariantRegistry)

	// routes
	Route() string
	NewHandler() sdk.Handler
	QuerierRoute() string
	NewQuerierHandler() sdk.Querier

	BeginBlock(sdk.Context, abci.RequestBeginBlock)
	EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
}
```

`AppModule`s are managed by the [module manager](#manager). This interface embeds the `AppModuleGenesis` interface so that the manager can access all the independent and genesis inter-dependent methods of the module. This means that a concrete type implementing the `AppModule` interface must either implement all the methods of `AppModuleGenesis` (and by extension `AppModuleBasic`), or include a concrete type that does as parameter. 

Let us go through the methods of `AppModule`:

- `RegisterInvariants(sdk.InvariantRegistry)`: Registers the [`invariants`](./invariants.md) of the module. If the invariants deviates from its predicted value, the [`InvariantRegistry`](./invariants.md#registry) triggers appropriate logic (most often the chain will be halted). 
- `Route()`: Returns the name of the module's route, for [`message`s](./messages-and-queries.md#messages) to be routed to the module by [`baseapp`](../core/baseapp.md#message-routing).
- `NewHandler()`: Returns a [`handler`](./handler.md) given the `Type()` of the `message`, in order to process the `message`. 
- `QuerierRoute()`: Returns the name of the module's query route, for [`queries`](./messages-and-queries.md#queries) to be routes to the module by [`baseapp`](../core/baseapp.md#query-routing). 
- `NewQuerierHandler()`: Returns a [`querier`](./querier.md) given the query `path`, in order to process the `query`. 
- `BeginBlock(sdk.Context, abci.RequestBeginBlock)`: This method gives module developers the option to implement logic that is automatically triggered at the beginning of each block. Implement empty if no logic needs to be triggered at the beginning of each block for this module. 
- `EndBlock(sdk.Context, abci.RequestEndBlock)`: This method gives module developers the option to implement logic that is automatically triggered at the beginning of each block. This is also where the module can inform the underlying consensus engine of validator set changes (e.g. the `staking` module). Implement empty if no logic needs to be triggered at the beginning of each block for this module. 

### Implementing the Application Module Interfaces

Typically, the various application module interfaces are implemented in a file called `module.go`, located in the module's folder (e.g. `./x/module/module.go`). 

Almost every module need to implement the `AppModuleBasic` and `AppModule` interfaces. If the module is only used for genesis, it will implement `AppModuleGenesis` instead of `AppModule`. The concrete type that implements the interface can add parameters that are required for the implementation of the various methods of the interface. For example, the `NewHandler()`  function often calls a `NewHandler(k keeper)` function defined in [`handler.go`](./handler.md) and therefore needs to pass the module's [`keeper`](./keeper.md) as parameter. 

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

The [`BasicManager`](https://github.com/cosmos/cosmos-sdk/blob/master/types/module/module.go#L60) is a structure that lists all the `AppModuleBasic` of an application: 

```go
type BasicManager map[string]AppModuleBasic
```

It implements the following methods:

- `NewBasicManager(modules ...AppModuleBasic)`: Constructor function. It takes a list of the application's `AppModuleBasic` and builds a new `BasicManager`. This function is generally called in the `init()` function of [`app.go`](../basics/app-anatomy.md#core-application-file) to quickly initialize the independent elements of the application's modules (click [here](https://github.com/cosmos/gaia/blob/master/app/app.go#L59-L74) to see an example).
- `RegisterCodec(cdc *codec.Codec)`: Registers the [`codec`s](../core/encoding.md) of each of the application's `AppModuleBasic`. This function is usually called early on in the [application's construction](../basics/app-anatomy.md#constructor).
- `DefaultGenesis()`: Provides default genesis information for modules in the application by calling the [`DefaultGenesis()`](./genesis.md#defaultgenesis) function of each module. It is used to construct a default genesis file for the application.
- `ValidateGenesis(genesis map[string]json.RawMessage)`: Validates the genesis information modules by calling the [`ValidateGenesis()`](./genesis.md#validategenesis) function of each module.
- `RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router)`: Registers REST routes for modules by calling the [`RegisterRESTRoutes`](./module-interfaces.md#register-routes) function of each module. This function is usually called function from the `main.go` function of the [application's command-line interface](../interfaces/cli.md).
- `AddTxCommands(rootTxCmd *cobra.Command, cdc *codec.Codec)`: Adds modules' transaction commands to the application's [`rootTxCommand`](../interfaces/cli.md#transaction-commands). This function is usually called function from the `main.go` function of the [application's command-line interface](../interfaces/cli.md).
- `AddQueryCommands(rootQueryCmd *cobra.Command, cdc *codec.Codec)`: Adds modules' query commands to the application's [`rootQueryCommand`](../interfaces/cli.md#query-commands). This function is usually called function from the `main.go` function of the [application's command-line interface](../interfaces/cli.md).


### `Manager`

The [`Manager`](https://github.com/cosmos/cosmos-sdk/blob/master/types/module/module.go#L203-L209) is a structure that lists all the `AppModule` of an application, and defines the order of execution between several key components of these modules:

```go
type Manager struct {
	Modules            map[string]AppModule
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
}
```

The module manager is used throughout the application whenever an action on a collection of module is required. It implements the following methods:

- `NewManager(modules ...AppModule)`: Constructor function. It takes a list of the application's `AppModule`s and builds a new `Manager`. It is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderInitGenesis(moduleNames ...string)`: Sets the order in which the [`InitGenesis`](./genesis.md#initgenesis) function of each module will be called when the application is first started. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderExportGenesis(moduleNames ...string)`: Sets the order in which the [`ExportGenesis`](./genesis.md#exportgenesis) function of each module will be called in case of an export. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderBeginBlockers(moduleNames ...string)`: Sets the order in which the `BeginBlock()` function of each module will be called at the beginning of each block. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `SetOrderEndBlockers(moduleNames ...string)`: Sets the order in which the `EndBlock()` function of each module will be called at the beginning of each block. This function is generally called from the application's main [constructor function](../basics/app-anatomy.md#constructor-function).
- `RegisterInvariants(ir sdk.InvariantRegistry)`: Registers the [invariants](./invariants.md) of each module.
- `RegisterRoutes(router sdk.Router, queryRouter sdk.QueryRouter)`: Registers module routes to the application's `router`, in order to route [`message`s](./messages-and-queries.md#messages) to the appropriate [`handler`](./handler.md), and module query routes to the application's `queryRouter`, in order to route [`queries`](./messages-and-queries.md#queries) to the appropriate [`querier`](./querier.md).
- `InitGenesis(ctx sdk.Context, genesisData map[string]json.RawMessage)`: Calls the [`InitGenesis`](./genesis.md#initgenesis) function of each module when the application is first started, in the order defined in `OrderInitGenesis`. Returns an `abci.ResponseInitChain` to the underlying consensus engine, which can contain validator updates. 
- `ExportGenesis(ctx sdk.Context)`: Calls the [`ExportGenesis`](./genesis.md#exportgenesis) function of each module, in the order defined in `OrderExportGenesis`. The export constructs a genesis file from a previously existing state, and is mainly used when a hard-fork upgrade of the chain is required. 
- `BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock)`: At the beginning of each block, this function calls the [`BeginBlock`] function of each module, in the order defined in `OrderBeginBlockers`. It creates a child [context](../core/context.md) with an event manager to aggregate [events](./events.md) emitted from all modules. The function returns an `abci.ResponseBeginBlock` which contains the aforementioned events. 
- `EndBlock(ctx sdk.Context, req abci.RequestEndBlock)`: At the end of each block, this function calls the [`EndBlock`] function of each module, in the order defined in `OrderEndBlockers`. It creates a child [context](../core/context.md) with an event manager to aggregate [events](./events.md) emitted from all modules. The function returns an `abci.ResponseEndBlock` which contains the aforementioned events, as well as validator set updates (if any).

## Next

Learn more about [`message`s and `queries`](./messages-and-queries.md).