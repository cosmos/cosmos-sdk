package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v2"
	v3 "github.com/cosmos/cosmos-sdk/x/bank/migrations/v3"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	return v2.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}

// Migrate2to3 migrates x/bank storage from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.MigrateStore(ctx, m.keeper.storeKey, m.keeper.cdc)
}

// Migrate3to4 migrates x/bank storage from version 3 to 4.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	oldParams := m.keeper.GetParams(ctx)
	m.keeper.SetAllSendEnabled(ctx, oldParams.GetSendEnabled())
	if err := m.keeper.SetParams(ctx, banktypes.NewParams(oldParams.DefaultSendEnabled)); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// TODO

	return nil
}
