package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - crisis genesis state
type GenesisState struct {
	ConstantFee sdk.Coin `json:"constant_fee"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(constantFee sdk.Coin) GenesisState {
	return GenesisState{
		ConstantFee: constantFee,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		ConstantFee: sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000)),
	}
}

// new crisis genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetConstantFee(ctx, data.ConstantFee)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	constantFee := keeper.GetConstantFee(ctx)
	return NewGenesisState(constantFee)
}

// ValidateGenesis - placeholder function
func ValidateGenesis(data GenesisState) error {
	return nil
}
