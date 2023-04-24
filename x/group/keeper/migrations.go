package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v3 "github.com/cosmos/cosmos-sdk/x/group/migrations/v3"
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
	// TODO

	// return v2.Migrate(
	// 	ctx,
	// 	m.keeper.key,
	// 	m.keeper.accKeeper,
	// 	m.keeper.groupPolicySeq,
	// 	m.keeper.groupPolicyTable,
	// )

	return nil
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate()
}
