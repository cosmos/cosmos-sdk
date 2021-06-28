package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v044 "github.com/cosmos/cosmos-sdk/x/params/legacy/v044"
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
	subspace, _ := m.keeper.GetSubspace(baseapp.Paramspace)
	return v044.MigrateStore(ctx, m.keeper.key, subspace)
}
