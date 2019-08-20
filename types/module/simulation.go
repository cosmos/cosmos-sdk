package module

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type SimulationManager struct {
	Modules       map[string]AppModule
	StoreDecoders sdk.StoreDecoderRegistry
	ParamChanges  []simulation.ParamChange
}

// NewSimulationManager creates a new SimulationManager object
func NewSimulationManager(moduleMap map[string]AppModule) *SimulationManager {
	return &SimulationManager{
		Modules:       moduleMap,
		StoreDecoders: make(sdk.StoreDecoderRegistry),
		ParamChanges:  []simulation.ParamChange{},
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
func (sm *SimulationManager) GenerateGenesisStates(input *GeneratorInput) {
	for _, module := range sm.Modules {
		module.GenerateGenesisState(input)
	}
}

// RandomizedSimParamChanges generates randomized contents for creating params change
// proposal transactions
func (sm *SimulationManager) RandomizedSimParamChanges(seed int64) {
	r := rand.New(rand.NewSource(seed))

	for _, module := range sm.Modules {
		sm.ParamChanges = append(sm.ParamChanges, module.RandomizedParams(r)...)
	}
}

// GeneratorInput is the input parameters used on each of the module's randomized
// GenesisState generator function
type GeneratorInput struct {
	AppParams    simulation.AppParams
	Cdc          *codec.Codec               // application codec
	R            *rand.Rand                 // random number
	GenState     map[string]json.RawMessage // genesis state
	Accounts     []simulation.Account       // simulation accounts
	InitialStake int64                      // initial coins per account
	NumBonded    int64                      // number of initially bonded acconts
	GenTimestamp time.Time                  // genesis timestamp
	UnbondTime   time.Duration              // TODO: consider removing due to non-determinism in map
}
