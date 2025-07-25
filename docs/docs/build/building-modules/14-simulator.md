---
sidebar_position: 1
---

# Module Simulation

:::note Pre-requisite Readings

* [Cosmos Blockchain Simulator](../../learn/advanced/12-simulation.md)

:::

## Synopsis

This document guides developers on integrating their custom modules with the Cosmos SDK `Simulations`.
Simulations are useful for testing edge cases in module implementations.

* [Simulation Package](#simulation-package)
* [Simulation App Module](#simulation-app-module)
* [SimsX](#simsx)
    * [Example Implementations](#example-implementations)
* [Store decoders](#store-decoders)
* [Randomized genesis](#randomized-genesis)
* [Random weighted operations](#random-weighted-operations)
    * [Using Simsx](#using-simsx)
* [App Simulator manager](#app-simulator-manager)
* [Running Simulations](#running-simulations)



## Simulation Package

The Cosmos SDK suggests organizing your simulation related code in a `x/<module>/simulation` package.

## Simulation App Module

To integrate with the Cosmos SDK `SimulationManager`, app modules must implement the `AppModuleSimulation` interface.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/3c6deab626648e47de752c33dac5d06af83e3ee3/types/module/simulation.go#L16-L27
```

See an example implementation of these methods from `x/distribution` [here](https://github.com/cosmos/cosmos-sdk/blob/b55b9e14fb792cc8075effb373be9d26327fddea/x/distribution/module.go#L170-L194).

## SimsX

Cosmos SDK v0.53.0 introduced a new package, `simsx`, providing improved DevX for writing simulation code.

It exposes the following extension interfaces that modules may implement to integrate with the new `simsx` runner.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/testutil/simsx/runner.go#L223-L234
```

These methods allow constructing randomized messages and/or proposal messages.

:::tip
Note that modules should **not** implement both `HasWeightedOperationsX` and `HasWeightedOperationsXWithProposals`.
See the runner code [here](https://github.com/cosmos/cosmos-sdk/blob/main/testutil/simsx/runner.go#L330-L339) for details

If the module does **not** have message handlers or governance proposal handlers, these interface methods do **not** need to be implemented.
:::

### Example Implementations

* `HasWeightedOperationsXWithProposals`: [x/gov](https://github.com/cosmos/cosmos-sdk/blob/main/x/gov/module.go#L242-L261)
* `HasWeightedOperationsX`: [x/bank](https://github.com/cosmos/cosmos-sdk/blob/main/x/bank/module.go#L199-L203)
* `HasProposalMsgsX`: [x/bank](https://github.com/cosmos/cosmos-sdk/blob/main/x/bank/module.go#L194-L197)

## Store decoders

Registering the store decoders is required for the `AppImportExport` simulation. This allows
for the key-value pairs from the stores to be decoded to their corresponding types.
In particular, it matches the key to a concrete type and then unmarshalls the value from the `KVPair` to the type provided.

Modules using [collections](https://github.com/cosmos/cosmos-sdk/blob/main/collections/README.md) can use the `NewStoreDecoderFuncFromCollectionsSchema` function that builds the decoder for you:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/x/bank/module.go#L181-L184
```

Modules not using collections must manually build the store decoder.
See the implementation [here](https://github.com/cosmos/cosmos-sdk/blob/main/x/distribution/simulation/decoder.go) from the distribution module for an example.

## Randomized genesis

The simulator tests different scenarios and values for genesis parameters.
App modules must implement a `GenerateGenesisState` method to generate the initial random `GenesisState` from a given seed.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/types/module/simulation.go#L20
```

See an example from `x/auth` [here](https://github.com/cosmos/cosmos-sdk/blob/main/x/auth/module.go#L169-L172).

Once the module's genesis parameters are generated randomly (or with the key and
values defined in a `params` file), they are marshaled to JSON format and added
to the app genesis JSON for the simulation.

## Random weighted operations

Operations are one of the crucial parts of the Cosmos SDK simulation. They are the transactions
(`Msg`) that are simulated with random field values. The sender of the operation
is also assigned randomly.

Operations on the simulation are simulated using the full [transaction cycle](../../learn/advanced/01-transactions.md) of a
`ABCI` application that exposes the `BaseApp`.

### Using Simsx

Simsx introduces the ability to define a `MsgFactory` for each of a module's messages.

These factories are registered in `WeightedOperationsX` and/or `ProposalMsgsX`.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/x/distribution/module.go#L196-L206
```

Note that the name passed in to `weights.Get` must match the name of the operation set in the `WeightedOperations`.

For example, if the module contains an operation `op_weight_msg_set_withdraw_address`, the name passed to `weights.Get` should be `msg_set_withdraw_address`.

See the `x/distribution` for an example of implementing message factories [here](https://github.com/cosmos/cosmos-sdk/blob/main/x/distribution/simulation/msg_factory.go)

## App Simulator manager

The following step is setting up the `SimulatorManager` at the app level. This
is required for the simulation test files in the next step.

```go
type CoolApp struct {
...
sm *module.SimulationManager
}
```

Within the constructor of the application, construct the simulation manager using the modules from `ModuleManager` and call the `RegisterStoreDecoders` method.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app.go#L650-L660
```

Note that you may override some modules.
This is useful if the existing module configuration in the `ModuleManager` should be different in the `SimulationManager`.

Finally, the application should expose the `SimulationManager` via the following method defined in the `Runtime` interface:

```go
// SimulationManager implements the SimulationApp interface
func (app *SimApp) SimulationManager() *module.SimulationManager {
return app.sm
}
```

## Running Simulations

To run the simulation, use the `simsx` runner.

Call the following function from the `simsx` package to begin simulating with a default seed:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/testutil/simsx/runner.go#L69-L88
```

If a custom seed is desired, tests should use `RunWithSeed`:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/b55b9e14fb792cc8075effb373be9d26327fddea/testutil/simsx/runner.go#L151-L168
```

These functions should be called in tests (i.e., app_test.go, app_sim_test.go, etc.)

Example:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/simapp/sim_test.go#L53-L65
```
