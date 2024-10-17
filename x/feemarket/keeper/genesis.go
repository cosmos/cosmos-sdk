package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/x/feemarket/types"
)

// InitGenesis initializes the feemarket module's state from a given genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, gs types.GenesisState) error {
	if err := gs.ValidateBasic(); err != nil {
		panic(err)
	}

	if gs.Params.Window != uint64(len(gs.State.Window)) {
		errors.New("genesis state and parameters do not match for window")
	}

	// Initialize the fee market state and parameters.
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}

	if err := k.SetState(ctx, gs.State); err != nil {
		return err
	}

	// always init enabled height to -1 until it is explicitly set later in the application
	return k.SetEnabledHeight(ctx, -1)
}

// ExportGenesis returns a GenesisState for a given context.
func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	// Get the feemarket module's parameters.
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// Get the feemarket module's state.
	state, err := k.GetState(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(params, state), nil
}
