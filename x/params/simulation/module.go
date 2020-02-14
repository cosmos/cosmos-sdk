package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sim "github.com/cosmos/cosmos-sdk/x/simulation"
)

var (
	_ module.AppModuleSimulation = AppModuleSimulation{}
)

// AppModuleSimulation implements an application module for the distribution module.
type AppModuleSimulation struct{}

// NewAppModule creates a new AppModuleSimulation object
func NewAppModule() AppModuleSimulation {
	return AppModuleSimulation{}
}

//____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState performs a no-op.
func (AppModuleSimulation) GenerateGenesisState(simState *module.SimulationState) {}

// ProposalContents returns all the params content functions used to
// simulate governance proposals.
func (am AppModuleSimulation) ProposalContents(simState module.SimulationState) []sim.WeightedProposalContent {
	return ProposalContents(simState.ParamChanges)
}

// RandomizedParams creates randomized distribution param changes for the simulator.
func (AppModuleSimulation) RandomizedParams(r *rand.Rand) []sim.ParamChange { return nil }

// RegisterStoreDecoder doesn't register any type.
func (AppModuleSimulation) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModuleSimulation) WeightedOperations(_ module.SimulationState) []sim.WeightedOperation {
	return nil
}
