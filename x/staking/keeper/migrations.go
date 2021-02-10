package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v042 "github.com/cosmos/cosmos-sdk/x/staking/legacy/v042"
)

// MigrationKeeper is an interface that the keeper implements for handling
// in-place store migrations.
type MigrationKeeper interface {
	// Migrate1 migrates the store from version 1 to 2.
	Migrate1(ctx sdk.Context) error
}

var _ MigrationKeeper = (*Keeper)(nil)

// Migrate1 implements MigrationKeeper.Migrate1 method.
func (keeper Keeper) Migrate1(ctx sdk.Context) error {
	return v042.MigrateStore(ctx, keeper.storeKey)
}
