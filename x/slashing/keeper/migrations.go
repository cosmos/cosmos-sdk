package keeper

import (
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v4"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate3to4 migrates the x/slashing module state from the consensus
// version 3 to version 4. Specifically, it migrates the validator missed block
// bitmap.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	store := runtime.KVStoreAdapter(m.keeper.storeService.OpenKVStore(ctx))
	params, err := m.keeper.Params.Get(ctx)
	if err != nil {
		return err
	}
	return v4.Migrate(ctx, m.keeper.cdc, store, params)
}
