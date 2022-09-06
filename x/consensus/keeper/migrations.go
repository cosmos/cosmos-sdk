package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	v2 "github.com/cosmos/cosmos-sdk/x/consensus/migrations/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper           *Keeper
	legacyParamStore exported.ParamStore
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper *Keeper, legacyParamStore exported.ParamStore) Migrator {
	return Migrator{
		keeper:           keeper,
		legacyParamStore: legacyParamStore,
	}
}

func (m Migrator) MigrateV1toV2(ctx sdk.Context) error {
	return v2.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc, m.legacyParamStore)
}
