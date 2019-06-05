package uniswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO

// GenesisState - uniswap genesis state
type GenesisState struct {
	NativeAssetDenom string
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(nativeAssetDenom string) GenesisState {
	return GenesisState{
		NativeAssetDenom: nativeAssetDenom,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		NativeAssetDenom: sdk.DefaultBondDenom,
	}
}

// InitGenesis new uniswap genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {

}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState(sdk.DefaultBondDenom)
}

// ValidateGenesis - placeholder function
func ValidateGenesis(data GenesisState) error {
	return nil
}
