package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v046 "github.com/cosmos/cosmos-sdk/x/params/migrations/v046"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}

// Migrate2to3 migrates x/staking state from consensus version 2 to 3.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v046.MigrateStore(ctx, m.keeper.Subspace(baseapp.Paramspace))
}
