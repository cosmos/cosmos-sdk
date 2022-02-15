package module

import (
	"encoding/json"

	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// AppModuleSimulation defines the standard functions that every module should expose
// for the SDK blockchain simulator
type AppModuleSimulation interface {
	// randomized genesis states
	GenerateGenesisState(input *SimulationState)

	// content functions used to simulate governance proposals
	ProposalContents(simState SimulationState) []simulation.WeightedProposalContent

	// randomized module parameters for param change proposals
	RandomizedParams(r *rand.Rand) []simulation.ParamChange

	// register a func to decode the each module's defined types from their corresponding store key
	RegisterStoreDecoder(sdk.StoreDecoderRegistry)

	// simulation operations (i.e msgs) with their respective weight
	WeightedOperations(simState SimulationState) []simulation.WeightedOperation
}

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type SimulationManager struct {
	Modules       []AppModuleSimulation    // array of app modules; we use an array for deterministic simulation tests
	StoreDecoders sdk.StoreDecoderRegistry // functions to decode the key-value pairs from each module's store
}

// NewSimulationManager creates a new SimulationManager object
//
// CONTRACT: All the modules provided must be also registered on the module Manager
func NewSimulationManager(modules ...AppModuleSimulation) *SimulationManager {
	return &SimulationManager{
		Modules:       modules,
		StoreDecoders: make(sdk.StoreDecoderRegistry),
	}
}

// GetProposalContents returns each module's proposal content generator function
// with their default operation weight and key.
func (sm *SimulationManager) GetProposalContents(simState SimulationState) []simulation.WeightedProposalContent {
	wContents := make([]simulation.WeightedProposalContent, 0, len(sm.Modules))
	for _, module := range sm.Modules {
		wContents = append(wContents, module.ProposalContents(simState)...)
	}

	return wContents
}

// RegisterStoreDecoders registers each of the modules' store decoders into a map
func (sm *SimulationManager) RegisterStoreDecoders() {
	for _, module := range sm.Modules {
		module.RegisterStoreDecoder(sm.StoreDecoders)
	}
}

// GenerateGenesisStates generates a randomized GenesisState for each of the
// registered modules
func (sm *SimulationManager) GenerateGenesisStates(simState *SimulationState) {
	for _, module := range sm.Modules {
		module.GenerateGenesisState(simState)
	}
}

// GenerateParamChanges generates randomized contents for creating params change
// proposal transactions
func (sm *SimulationManager) GenerateParamChanges(seed int64) (paramChanges []simulation.ParamChange) {
	r := rand.New(rand.NewSource(seed))

	for _, module := range sm.Modules {
		paramChanges = append(paramChanges, module.RandomizedParams(r)...)
	}

	return
}

// WeightedOperations returns all the modules' weighted operations of an application
func (sm *SimulationManager) WeightedOperations(simState SimulationState) []simulation.WeightedOperation {
	wOps := make([]simulation.WeightedOperation, 0, len(sm.Modules))
	for _, module := range sm.Modules {
		wOps = append(wOps, module.WeightedOperations(simState)...)
	}

	return wOps
}

// SimulationState is the input parameters used on each of the module's randomized
// GenesisState generator function
type SimulationState struct {
	AppParams    simulation.AppParams
	Cdc          codec.JSONCodec                      // application codec
	Rand         *rand.Rand                           // random number
	GenState     map[string]json.RawMessage           // genesis state
	Accounts     []simulation.Account                 // simulation accounts
	InitialStake int64                                // initial coins per account
	NumBonded    int64                                // number of initially bonded accounts
	GenTimestamp time.Time                            // genesis timestamp
	UnbondTime   time.Duration                        // staking unbond time stored to use it as the slashing maximum evidence duration
	ParamChanges []simulation.ParamChange             // simulated parameter changes from modules
	Contents     []simulation.WeightedProposalContent // proposal content generator functions with their default weight and app sim key
}
