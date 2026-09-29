package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v6 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v6"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator instance.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}

// Migrate5to6 migrates the x/staking module state from consensus version 5 to
// version 6. Specifically, it adds the KeyRotationFee param.
func (m Migrator) Migrate5to6(ctx sdk.Context) error {
	return v6.Migrate(ctx, m.keeper)
}
