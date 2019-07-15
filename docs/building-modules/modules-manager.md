# Module Manager and `AppModule` Interface

## Pre-requisite Reading

- [Introduction to SDK Modules](./intro.md)

## Synopsis

Cosmos SDK modules need to implement the [`AppModule` interfaces](#application-module-interfaces), in order to be managed by the application's [module manager](#module-manager). The module manager plays an important role in [`message` and `query` routing](../core/baseapp.md#routing), and allows the application developer to set the order of execution of a variety of functions like [`BeginBlocker` and `EndBlocker`](../basics/app-anatomy.md#begingblocker-and-endblocker). 

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

### `AppModule`

### `AppModuleGenesis`

### Implementing the Application Module Interfaces

## Module Manager

