package upgrade

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - upgrade genesis state
type GenesisState struct {
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState() GenesisState {
	return GenesisState{}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// new crisis genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState()
}

// ValidateGenesis - placeholder function
func ValidateGenesis(data GenesisState) error {
	return nil
}
