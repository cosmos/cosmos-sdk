package keeper

import (
	v2 "cosmossdk.io/x/group/migrations/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(
		ctx,
		m.keeper.key,
		m.keeper.accKeeper,
		m.keeper.groupPolicySeq,
		m.keeper.groupPolicyTable,
	)
}
