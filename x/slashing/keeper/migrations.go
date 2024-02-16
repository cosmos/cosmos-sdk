package keeper

import (
	"context"

	v4 "cosmossdk.io/x/slashing/migrations/v4"

	"github.com/cosmos/cosmos-sdk/runtime"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx context.Context) error {
	return nil
}

// Migrate2to3 migrates the x/slashing module state from the consensus
// version 2 to version 3. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/slashing
// module state.
func (m Migrator) Migrate2to3(ctx context.Context) error {
	return nil
}

// Migrate3to4 migrates the x/slashing module state from the consensus
// version 3 to version 4. Specifically, it migrates the validator missed block
// bitmap.
func (m Migrator) Migrate3to4(ctx context.Context) error {
	store := runtime.KVStoreAdapter(m.keeper.environment.KVStoreService.OpenKVStore(ctx))
	params, err := m.keeper.Params.Get(ctx)
	if err != nil {
		return err
	}
	return v4.Migrate(ctx, m.keeper.cdc, store, params)
}
