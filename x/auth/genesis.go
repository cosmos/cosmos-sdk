package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	CollectedFees sdk.Coins `json:"collected_fees"` // collected fees
}

// Create a new genesis state
func NewGenesisState(collectedFees sdk.Coins) GenesisState {
	return GenesisState{
		CollectedFees: collectedFees,
	}
}

// Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(sdk.Coins{})
}

// Init store state from genesis data
func InitGenesis(ctx sdk.Context, keeper FeeCollectionKeeper, data GenesisState) {
	keeper.setCollectedFees(ctx, data.CollectedFees)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, keeper FeeCollectionKeeper) GenesisState {
	collectedFees := keeper.GetCollectedFees(ctx)
	return NewGenesisState(collectedFees)
}
