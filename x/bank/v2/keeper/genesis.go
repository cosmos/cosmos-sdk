package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/x/bank/v2/types"
)

// InitGenesis initializes the bank/v2 module genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	if err := k.params.Set(ctx, state.Params); err != nil {
		return fmt.Errorf("failed to set params: %v", err)
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %v", err)
	}

	return types.NewGenesisState(params), nil
}
