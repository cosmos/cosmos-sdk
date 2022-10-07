package keeper

import (
	"encoding/binary"

	store2 "github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
	return migrateDoneUpgradeKeys(ctx, m.keeper.storeKey)
}

func migrateDoneUpgradeKeys(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)
	newStore := store2.NewStoreAPI(store)
	oldDoneStore := prefix.NewStore(store, []byte{types.DoneByte})
	oldDoneStore2 := store2.NewStoreAPI(oldDoneStore)
	oldDoneStoreIter := oldDoneStore.Iterator(nil, nil)
	defer oldDoneStoreIter.Close()

	for ; oldDoneStoreIter.Valid(); oldDoneStoreIter.Next() {
		oldKey := oldDoneStoreIter.Key()
		upgradeName := string(oldKey)
		upgradeHeight := int64(binary.BigEndian.Uint64(oldDoneStoreIter.Value()))
		newKey := encodeDoneKey(upgradeName, upgradeHeight)

		newStore.Set(newKey, []byte{1})
		oldDoneStore2.Delete(oldKey)
	}
	return nil
}
