package module

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// AppModuleSimulation defines the standard functions that every module should expose
// for the SDK blockchain simulator
type AppModuleSimulation interface {
	// register a func to decode the each module's defined types from their corresponding store key
	RegisterStoreDecoder(sdk.StoreDecoderRegistry)

	// randomized genesis states
	GenerateGenesisState(input *SimulationState)

	// randomized module parameters for param change proposals
	RandomizedParams(r *rand.Rand) []simulation.ParamChange
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

// RegisterStoreDecoders registers each of the modules' store decoders into a map
func (sm *SimulationManager) RegisterStoreDecoders() {
	for _, module := range sm.Modules {
		module.RegisterStoreDecoder(sm.StoreDecoders)
	}
}

// GenerateGenesisStates generates a randomized GenesisState for each of the
// registered modules
func (sm *SimulationManager) GenerateGenesisStates(input *SimulationState) {
	for _, module := range sm.Modules {
		module.GenerateGenesisState(input)
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

// SimulationState is the input parameters used on each of the module's randomized
// GenesisState generator function
type SimulationState struct {
	AppParams    simulation.AppParams
	Cdc          *codec.Codec               // application codec
	Rand         *rand.Rand                 // random number
	GenState     map[string]json.RawMessage // genesis state
	Accounts     []simulation.Account       // simulation accounts
	InitialStake int64                      // initial coins per account
	NumBonded    int64                      // number of initially bonded acconts
	GenTimestamp time.Time                  // genesis timestamp
	UnbondTime   time.Duration              // staking unbond time stored to use it as the slashing maximum evidence duration
}
