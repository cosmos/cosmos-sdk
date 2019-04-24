package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// InitGenesis new mint genesis
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.SetMinter(ctx, data.Minter)
	k.SetParams(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {

	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)
	return types.NewGenesisState(minter, params)
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data types.GenesisState) error {
	err := types.ValidateParams(data.Params)
	if err != nil {
		return err
	}

	err = types.ValidateMinter(data.Minter)
	if err != nil {
		return err
	}

	return nil
}
