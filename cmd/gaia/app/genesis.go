package app

import (
	"encoding/json"
)

// Genesis State of the blockchain
type GenesisState struct {
	Modules map[string]json.RawMessage `json:"modules"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(modules map[string]json.RawMessage) GenesisState {
	return GenesisState{
		Modules: modules,
	}
}

// NewDefaultGenesisState generates the default state for gaia.
func NewDefaultGenesisState() GenesisState {
	return NewGenesisState(nil, mbm.DefaultGenesis(), nil)
}
