package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	orm "github.com/cosmos/cosmos-sdk/x/group/migrations/legacyorm"
	v2 "github.com/cosmos/cosmos-sdk/x/group/migrations/v2"
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
	groupPolicyTable, err := orm.NewPrimaryKeyTable([2]byte{v2.GroupPolicyTablePrefix}, &v2.GroupPolicyInfo{}, m.keeper.cdc)
	if err != nil {
		return err
	}

	return v2.Migrate(
		ctx,
		// m.keeper.storeService, // TODO
		nil,
		m.keeper.accKeeper,
		orm.NewSequence(v2.GroupPolicyTableSeqPrefix),
		*groupPolicyTable,
	)
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v3.Migrate()
}
