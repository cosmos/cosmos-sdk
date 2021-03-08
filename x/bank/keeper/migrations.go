package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042 "github.com/cosmos/cosmos-sdk/x/bank/legacy/v042"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper BaseKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper BaseKeeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v042.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}
