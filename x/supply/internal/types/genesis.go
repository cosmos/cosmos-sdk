package types

import (
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// GenesisState is the supply state that must be provided at genesis.
type GenesisState struct {
	Supply exported.SupplyI `json:"supply" yaml:"supply"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(supply exported.SupplyI) GenesisState {
	return GenesisState{supply}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultSupply())
}
