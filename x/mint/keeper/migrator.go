package keeper

import (
	"context"

	v4 "cosmossdk.io/x/mint/migrations/v4"
	"cosmossdk.io/x/mint/types"
)

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns Migrator instance for the state migration.
func NewMigrator(k Keeper) Migrator {
	return Migrator{
		keeper: k,
	}
}

// Migrate1to2 migrates the x/mint module state from the consensus version 1 to
// version 2. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/mint
// module state.
func (m Migrator) Migrate1to2(ctx context.Context) error {
	return nil
}

// Migrate2to3 migrates the x/mint module state from the consensus version 2 to
// version 3.
func (m Migrator) Migrate2to3(ctx context.Context) error {
	params, err := m.keeper.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Initialize the new MaxSupply parameter with the default value
	defaultParams := types.DefaultParams()
	params.MaxSupply = defaultParams.MaxSupply

	// Set the updated params
	err = m.keeper.Params.Set(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

// Migrate3to4 migrates from version 3 to 4.
func (m Migrator) Migrate3to4(ctx context.Context) error {
	// Initialize the new LastReductionEpoch with the default value
	defaultGenesisState := types.DefaultGenesisState()
	reductionStartedEpoch := defaultGenesisState.ReductionStartedEpoch
	err := m.keeper.LastReductionEpoch.Set(ctx, reductionStartedEpoch)
	if err != nil {
		return err
	}
	return v4.MigrateStore(ctx, m.keeper.LastReductionEpoch)
}
