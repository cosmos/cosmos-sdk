package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - minter state
type GenesisState struct {
	Minter Minter `json:"minter"` // minter object
	Params Params `json:"params"` // inflation params
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params) GenesisState {
	return GenesisState{
		Minter: minter,
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Minter: DefaultInitialMinter(),
		Params: DefaultParams(),
	}
}

// new mint genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.SetMinter(ctx, data.Minter)
	if err := keeper.SetParams(ctx, data.Params); err != nil {
		// TODO: return error - needs rewrite interfaces
		// and handle error on the caller side
		// check PR #3782
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, and validator/delegator distribution info's
func ExportGenesis(ctx sdk.Context, keeper Keeper) (GenesisState, error) {
	minter, err := keeper.GetMinter(ctx)
	if err != nil {
		return GenesisState{}, err
	}
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return GenesisState{}, err
	}
	return NewGenesisState(minter, params), nil
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
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
