package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// LegacyProtocolVersionByte was the prefix to look up Protocol Version (AppVersion)
	LegacyProtocolVersionByte = 0x3
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return migrateDoneUpgradeKeys(ctx, m.keeper.storeKey)
}

func migrateDoneUpgradeKeys(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)
	oldDoneStore := prefix.NewStore(store, []byte{types.DoneByte})
	oldDoneStoreIter := oldDoneStore.Iterator(nil, nil)
	defer oldDoneStoreIter.Close()

	for ; oldDoneStoreIter.Valid(); oldDoneStoreIter.Next() {
		oldKey := oldDoneStoreIter.Key()
		upgradeName := string(oldKey)
		upgradeHeight := int64(binary.BigEndian.Uint64(oldDoneStoreIter.Value()))
		newKey := encodeDoneKey(upgradeName, upgradeHeight)

		store.Set(newKey, []byte{1})
		oldDoneStore.Delete(oldKey)
	}
	return nil
}

func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return migrateAppVersion(ctx, m.keeper)
}

func migrateAppVersion(ctx sdk.Context, keeper *Keeper) error {
	if keeper.versionModifier == nil {
		return fmt.Errorf("version modifier is not set")
	}
	store := ctx.KVStore(keeper.storeKey)
	// if the key was never set then we don't need to
	if !store.Has([]byte{LegacyProtocolVersionByte}) {
		return nil
	}
	versionBytes := store.Get([]byte{LegacyProtocolVersionByte})
	appVersion := binary.BigEndian.Uint64(versionBytes)

	if err := keeper.versionModifier.SetAppVersion(ctx, appVersion); err != nil {
		return fmt.Errorf("error migration app version: %w", err)
	}

	store.Delete([]byte{LegacyProtocolVersionByte})
	return nil
}
