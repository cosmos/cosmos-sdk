<!--
order: 9
-->

# Module Genesis

Modules generally handle a subset of the state and, as such, they need to define the related subset of the genesis file as well as methods to initialize, verify and export it. {synopsis}

## Pre-requisite Readings

- [Module Manager](./module-manager.md) {prereq}
- [Keepers](./keeper.md) {prereq}

## Type Definition 

The subset of the genesis state defined from a given module is generally defined in a `genesis.proto` file ([more info](../core/encoding.md#gogoproto) on how to define protobuf messages). The struct defining the module's subset of the genesis state is usually called `GenesisState` and contains all the module-related values that need to be initialized during the genesis process. 

See an example of `GenesisState` protobuf message definition from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/a9547b54ffac9729fe1393651126ddfc0d236cff/proto/cosmos/auth/v1beta1/genesis.proto

Next we present the main genesis-related methods that need to be implemented by module developers in order for their module to be used in Cosmos SDK applications. 

### `DefaultGenesis`

The `DefaultGenesis()` method is a simple method that calls the constructor function for `GenesisState` with the default value for each parameter. See an example from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/module.go#L48-L52

### `ValidateGenesis`

The `ValidateGenesis(genesisState GenesisState)` method is called to verify that the provided `genesisState` is correct. It should perform validity checks on each of the parameter listed in `GenesisState`. See an example from the `auth` module:

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/types/genesis.go#L57-L70

## Other Genesis Methods

Other than the methods related directly to `GenesisState`, module developers are expected to implement two other methods as part of the [`AppModuleGenesis` interface](./module-manager.md#appmodulegenesis) (only if the module needs to initialize a subset of state in genesis). These methods are [`InitGenesis`](#initgenesis) and [`ExportGenesis`](#exportgenesis).

### `InitGenesis`

The `InitGenesis` method is executed during [`InitChain`](../core/baseapp.md#initchain) when the application is first started. Given a `GenesisState`, it initializes the subset of the state managed by the module by using the module's [`keeper`](./keeper.md) setter function on each parameter within the `GenesisState`. 

The [module manager](./module-manager.md#manager) of the application is responsible for calling the `InitGenesis` method of each of the application's modules, in order. This order is set by the application developer via the manager's `SetOrderGenesisMethod`, which is called in the [application's constructor function](../basics/app-anatomy.md#constructor-function)

See an example of `InitGenesis` from the `auth` module

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/genesis.go#L13-L28

### `ExportGenesis`

The `ExportGenesis` method is executed whenever an export of the state is made. It takes the latest known version of the subset of the state managed by the module and creates a new `GenesisState` out of it. This is mainly used when the chain needs to be upgraded via a hard fork. 

See an example of `ExportGenesis` from the `auth` module.

+++ https://github.com/cosmos/cosmos-sdk/blob/64b6bb5270e1a3b688c2d98a8f481ae04bb713ca/x/auth/genesis.go#L31-L42

## Next {hide}

Learn about [modules interfaces](module-interfaces.md) {hide}