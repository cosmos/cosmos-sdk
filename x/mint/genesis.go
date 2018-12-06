package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all distribution state that must be provided at genesis
type GenesisState struct {
	Minter Minter `json:"minter"` // minter object
	Params Params `json:"params"` // inflation params
}

func NewGenesisState(minter Minter, params Params) GenesisState {
	return GenesisState{
		Minter: minter,
		Params: params,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter: DefaultInitialMinter(),
		Params: DefaultParams(),
	}
}

// new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	keeper.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, and validator/delegator distribution info's
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {

	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)
	return NewGenesisState(minter, params)
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func ValidateGenesis(data GenesisState) error {
	err := validateParams(data.Params)
	if err != nil {
		return err
	}
	err = validateMinter(data.Minter)
	if err != nil {
		return err
	}
	return nil
}
