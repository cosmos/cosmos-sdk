package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SimulationManager defines a simulation manager that provides the high level utility
// for managing and executing simulation functionalities for a group of modules
type SimulationManager struct {
	Modules       map[string]AppModule
	StoreDecoders sdk.StoreDecoderRegistry
}

// NewSimulationManager creates a new SimulationManager object
func NewSimulationManager(moduleMap map[string]AppModule) *SimulationManager {
	return &SimulationManager{
		Modules:       moduleMap,
		StoreDecoders: make(sdk.StoreDecoderRegistry),
	}
}

// RegisterStoreDecoders registers each of the modules' store decoders into a map
func (sm *SimulationManager) RegisterStoreDecoders() {
	for _, module := range sm.Modules {
		module.RegisterStoreDecoder(sm.StoreDecoders)
	}
}
