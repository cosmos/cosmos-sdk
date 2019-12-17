<!--
order: 9
synopsis: Modules generally handle a subset of the state and, as such, they need to define the related subset of the genesis file as well as methods to initialize, verify and export it. 
-->

# Module Genesis

## Pre-requisite Readings {hide}

- [Module Manager](./module-manager.md) {prereq}
- [Keepers](./keeper.md) {prereq}

## Type Definition 

The subset of the genesis state defined from a given module is generally defined in a `./internal/types/genesis.go` file, along with the `DefaultGenesis` and `ValidateGenesis` methods. The struct defining the module's subset of the genesis state is usually called `GenesisState` and contains all the module-related values that need to be initialized during the genesis process. 

See an example of `GenesisState` type definition from the nameservice tutorial

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/genesis.go#L10-L12

Next we present the main genesis-related methods that need to be implemented by module developers in order for their module to be used in Cosmos SDK applications. 

### `DefaultGenesis`

The `DefaultGenesis()` method is a simple method that calls the constructor function for `GenesisState` with the default value for each parameter. See an example from the `nameservice` module:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/genesis.go#L33-L37

### `ValidateGenesis`

The `ValidateGenesis(genesisState GenesisState)` method is called to verify that the provided `genesisState` is correct. It should perform validity checks on each of the parameter listed in `GenesisState`. See an example from the `nameservice` module:

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/genesis.go#L18-L31

## Other Genesis Methods

Other than the methods related directly to `GenesisState`, module developers are expected to implement two other methods as part of the [`AppModuleGenesis` interface](./module-manager.md#appmodulegenesis) (only if the module needs to initialize a subset of state in genesis). These methods are [`InitGenesis`](#initgenesis) and [`ExportGenesis`](#exportgenesis).

### `InitGenesis`

The `InitGenesis` method is executed during [`InitChain`](../core/baseapp.md#initchain) when the application is first started. Given a `GenesisState`, it initializes the subset of the state managed by the module by using the module's [`keeper`](./keeper.md) setter function on each parameter within the `GenesisState`. 

The [module manager](./module-manager.md#manager) of the application is responsible for calling the `InitGenesis` method of each of the application's modules, in order. This order is set by the application developer via the manager's `SetOrderGenesisMethod`, which is called in the [application's constructor function](../basics/app-anatomy.md#constructor-function)

See an example of `InitGenesis` from the nameservice tutorial

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/genesis.go#L39-L44

### `ExportGenesis`

The `ExportGenesis` method is executed whenever an export of the state is made. It takes the latest known version of the subset of the state managed by the module and creates a new `GenesisState` out of it. This is mainly used when the chain needs to be upgraded via a hard fork. 

See an example of `ExportGenesis` from the nameservice tutorial.

+++ https://github.com/cosmos/sdk-tutorials/blob/86a27321cf89cc637581762e953d0c07f8c78ece/nameservice/x/nameservice/genesis.go#L46-L57

## Next {hide}

Learn about [modules interfaces](#module-interfaces.md) {hide}