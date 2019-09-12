package vesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis - Init store state from genesis data
func InitGenesis(ctx sdk.Context, data GenesisState) {
	return
}

// ExportGenesis returns a GenesisState for a given context
func ExportGenesis(ctx sdk.Context) GenesisState {
	return NewGenesisState()
}
