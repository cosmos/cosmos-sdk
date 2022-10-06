package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	vestingKeeper VestingKeeper
	accountKeeper types.AccountKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(vk VestingKeeper, ak types.AccountKeeper) Migrator {
	return Migrator{vestingKeeper: vk, accountKeeper: ak}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	m.accountKeeper.IterateAccounts(ctx, func(account authtypes.AccountI) bool {
		if va, ok := account.(exported.VestingAccount); ok {
			m.vestingKeeper.AddVestingAccount(ctx, va.GetAddress())
		}
		return false
	})

	return nil
}
