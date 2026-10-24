package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// InitGenesis validates the genesis state and writes it to the store.
func (k Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
	if err := gs.Validate(k.addressCodec); err != nil {
		return err
	}
	if err := k.Params.Set(ctx, gs.Params); err != nil {
		return err
	}
	for _, p := range gs.DisabledPatterns {
		if err := k.DisabledList.Set(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis returns the module state as a genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var disabled []string
	err = k.DisabledList.Walk(ctx, nil, func(p string) (bool, error) {
		disabled = append(disabled, p)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(params, disabled), nil
}
