package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the supply state that must be provided at genesis.
type GenesisState struct {
	Supply sdk.Coins `json:"supply" yaml:"supply"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(supply sdk.Coins) GenesisState {
	return GenesisState{supply}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultSupply().GetTotal())
}
