# Module Simulation

## Prerequisites

* [Cosmos Blockchain Simulator](../concepts/using-the-sdk/simulation.md)
<!-- Application SimulationManager -->

## Synopsis

This document details how to define each module simulation functions to be
integrated with the application `SimulationManager`.

* [Simulation Package](#cli)
  * [Type decoders](#type-decoders)
  * [Randomized genesis](#randomized-genesis)
  * [Randomized parameters](#randomized-parameters)
  * [Random weighted operations](#random-weighted-operations)
  * [Random proposal contents](#random-proposal-contents)
* [Registering simulation functions](#registering-simulation-functions)

## Simulation package

### Type decoders

### Randomized genesis

### Randomized parameters

### Random weighted operations

### Random proposal contents

## Registering simulation functions

Now that all the required functions are defined, we need to integrate them into the module pattern:

```go
// x/<module>/module.go

import (
  // ...
  "github.com/cosmos/cosmos-sdk/x/<module>/simulation"
  sim "github.com/cosmos/cosmos-sdk/x/simulation"
)

var (
  _ module.AppModule           = AppModule{}
  _ module.AppModuleBasic      = AppModuleBasic{}
  _ module.AppModuleSimulation = AppModule{} //  <-- AppModule now implements the module.AppModuleSimulation interface
)

// ...
//____________________________________________________________________________
// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the <module> module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
 simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the <module> content functions used to
// simulate governance proposals.
//
// NOTE: return nil if your module doesn't integrate the goverance module
func (am AppModule) ProposalContents(simState module.SimulationState) []sim.WeightedProposalContent {
 return simulation.ProposalContents(simState,
   // keepers defined in AppModule
 )
}

// RandomizedParams creates randomized <module> param changes for the simulator.
//
// NOTE: return nil if your module doesn't integrate the params module
func (AppModule) RandomizedParams(r *rand.Rand) []sim.ParamChange {
 return simulation.ParamChanges(r)
}

// RegisterStoreDecoder registers a decoder for <module> module's types
func (AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {
  sdr[StoreKey] = simulation.DecodeStore
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []sim.WeightedOperation {
 return simulation.WeightedOperations(
   simState.AppParams, simState.Cdc,
   // keepers defined in AppModule
   )
}
```
