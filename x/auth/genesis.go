package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	CollectedFees sdk.Coins `json:"collected_fees"` // collected fees
	Params        Params    `json:"params"`
}

// Create a new genesis state
func NewGenesisState(collectedFees sdk.Coins, params Params) GenesisState {
	return GenesisState{
		CollectedFees: collectedFees,
		Params:        params,
	}
}

// Return a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(sdk.Coins{}, DefaultParams())
}

// Init store state from genesis data
func InitGenesis(ctx sdk.Context, ak AccountKeeper, fck FeeCollectionKeeper, data GenesisState) {
	ak.SetParams(ctx, data.Params)
	fck.setCollectedFees(ctx, data.CollectedFees)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak AccountKeeper, fck FeeCollectionKeeper) GenesisState {
	collectedFees := fck.GetCollectedFees(ctx)
	params := ak.GetParams(ctx)

	return NewGenesisState(collectedFees, params)
}
