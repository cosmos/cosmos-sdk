package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Migrator handles in-place store migrations for x/bank.
type Migrator struct {
	keeper BaseKeeper
}

// NewMigrator returns a new bank Migrator.
func NewMigrator(k BaseKeeper) Migrator {
	return Migrator{keeper: k}
}

// Migrate4to5 rewrites legacy Coin balances so DenomOwners can see them.
func (m Migrator) Migrate4to5(ctx sdk.Context) error {
	_, err := m.keeper.MigrateLegacyBalances(ctx)
	return err
}
