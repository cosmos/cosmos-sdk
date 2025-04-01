package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
)

// InitGenesis sets epoch info from genesis
func (k *Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) error {
	for _, epoch := range genState.Epochs {
		err := k.AddEpochInfo(ctx, epoch)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis returns the capability module's exported genesis.
func (k *Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	genesis := types.DefaultGenesis()
	epochs, err := k.AllEpochInfos(ctx)
	if err != nil {
		return nil, err
	}
	genesis.Epochs = epochs
	return genesis, nil
}
