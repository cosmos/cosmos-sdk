package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3 "github.com/cosmos/cosmos-sdk/x/mint/migrations/v3"
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

// Migrate2to3 migrates the x/mint module state from the consensus version 2 to
// version 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate(ctx, m.keeper.storeService.OpenKVStore(ctx), m.keeper.cdc, m.keeper.Params)
}
