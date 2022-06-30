package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	v043 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v043"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
	legacySubspace paramstypes.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, ss paramstypes.Subspace) Migrator {
	return Migrator{keeper: keeper, legacySubspace: ss}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v043.MigrateStore(ctx, m.keeper.storeKey)
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	var currParams types.Params
	m.legacySubspace.GetParamSet(ctx, &currParams)
	return m.keeper.SetParams(ctx, currParams)
}
