package module

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type SimulationManager struct {
	Modules       []AppModule              // array of app modules; we use an array for deterministic simulation tests
	StoreDecoders sdk.StoreDecoderRegistry // functions to decode the key-value pairs from each module's store
	ParamChanges  []simulation.ParamChange // list of parameters changes transactions run by the simulator
}

// NewSimulationManager creates a new SimulationManager object
func NewSimulationManager(moduleMap map[string]AppModule, moduleNames []string) *SimulationManager {
	var modules []AppModule

	for _, name := range moduleNames {
		module, ok := moduleMap[name]
		if !ok {
			panic(fmt.Sprintf("module %s is not registered on the module manager", name))
		}
		modules = append(modules, module)
	}

	return &SimulationManager{
		Modules:       modules,
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
func (sm *SimulationManager) GenerateGenesisStates(input *SimulationState) {
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
