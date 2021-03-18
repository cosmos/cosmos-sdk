# Module Simulation

## Prerequisites

* [Cosmos Blockchain Simulator](./../using-the-sdk/simulation.md)

## Synopsis

This document details how to define each module simulation functions to be
integrated with the application `SimulationManager`.

* [Simulation package](#simulation-package)
  * [Store decoders](#store-decoders)
  * [Randomized genesis](#randomized-genesis)
  * [Randomized parameters](#randomized-parameters)
  * [Random weighted operations](#random-weighted-operations)
  * [Random proposal contents](#random-proposal-contents)
* [Registering the module simulation functions](#registering-simulation-functions)
* [App simulator manager](#app-simulator-manager)
* [Simulation tests](#simulation-tests)

## Simulation package

Every module that implements the SDK simulator needs to have a `x/<module>/simulation`
package which contains the primary functions required by the fuzz tests: store
decoders, randomized genesis state and parameters, weighted operations and proposal
contents.

### Store decoders

Registering the store decoders is required for the `AppImportExport`. This allows
for the key-value pairs from the stores to be decoded (_i.e_ unmarshalled)
to their corresponding types. In particular, it matches the key to a concrete type
and then unmarshals the value from the `KVPair` to the type provided.

You can use the example [here](https://github.com/cosmos/cosmos-sdk/blob/v0.42.0/x/distribution/simulation/decoder.go) from the distribution module to implement your store decoders.

### Randomized genesis

The simulator tests different scenarios and values for genesis parameters
in order to fully test the edge cases of specific modules. The `simulator` package from each module must expose a `RandomizedGenState` function to generate the initial random `GenesisState` from a given seed.

Once the module genesis parameter are generated randomly (or with the key and
values defined in a `params` file), they are marshaled to JSON format and added
to the app genesis JSON to use it on the simulations.

You can check an example on how to create the randomized genesis [here](https://github.com/cosmos/cosmos-sdk/blob/v0.42.0/x/staking/simulation/genesis.go).

### Randomized parameter changes

The simulator is able to test parameter changes at random. The simulator package from each module must contain a `RandomizedParams` func that will simulate parameter changes of the module throughout the simulations lifespan.

You can see how an example of what is needed to fully test parameter changes [here](https://github.com/cosmos/cosmos-sdk/blob/v0.42.0/x/staking/simulation/params.go)

### Random weighted operations

Operations are one of the crucial parts of the SDK simulation. They are the transactions
(`Msg`) that are simulated with random field values. The sender of the operation
is also assigned randomly.

Operations on the simulation are simulated using the full [transaction cycle](../core/transactions.md) of a
`ABCI` application that exposes the `BaseApp`.

Shown below is how weights are set:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/x/staking/simulation/operations.go#L18-L68

As you can see, the weights are predefined in this case. Options exist to override this behavior with different weights. One option is to use `*rand.Rand` to define a random weight for the operation, or you can inject your own predefined weights.

Here is how one can override the above package `simappparams`.

+++ https://github.com/cosmos/gaia/blob/master/sims.mk#L9-L22

For the last test a tool called runsim  <!-- # TODO: add link to runsim readme when its created --> is used, this is used to parallelize go test instances, provide info to Github and slack integrations to provide information to your team on how the simulations are running.  

### Random proposal contents

Randomized governance proposals are also supported on the SDK simulator. Each
module must define the governance proposal `Content`s that they expose and register
them to be used on the parameters.

## Registering simulation functions

Now that all the required functions are defined, we need to integrate them into the module pattern within the `module.go`:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.0/x/distribution/module.go

## App Simulator manager

The following step is setting up the `SimulatorManager` at the app level. This
is required for the simulation test files on the next step.

```go
type CustomApp struct {
  ...
  sm *module.SimulationManager
}
```

Then at the instantiation of the application, we create the `SimulationManager`
instance in the same way we create the `ModuleManager` but this time we only pass
the modules that implement the simulation functions from the `AppModuleSimulation`
interface described above.

```go
func NewCustomApp(...) {
  // create the simulation manager and define the order of the modules for deterministic simulations
  app.sm = module.NewSimulationManager(
    auth.NewAppModule(app.accountKeeper),
    bank.NewAppModule(app.bankKeeper, app.accountKeeper),
    supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
    ov.NewAppModule(app.govKeeper, app.accountKeeper, app.supplyKeeper),
    mint.NewAppModule(app.mintKeeper),
    distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.supplyKeeper, app.stakingKeeper),
    staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),
    slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
  )

  // register the store decoders for simulation tests
  app.sm.RegisterStoreDecoders()
  ...
}
```
