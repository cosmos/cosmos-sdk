package keeper

import (
	"context"

	"cosmossdk.io/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := keeper.Minter.Set(ctx, data.Minter); err != nil {
		return err
	}

	return keeper.Params.Set(ctx, data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	minter, err := keeper.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(minter, params), nil
}
