package types

import (
	corestore "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
)

// UpgradeStoreLoader is used to prepare baseapp with a fixed StoreLoader
// pattern. This is useful for custom upgrade loading logic.
func UpgradeStoreLoader(upgradeHeight int64, storeUpgrades *corestore.StoreUpgrades) baseapp.StoreLoader {
	return func(ms storetypes.CommitMultiStore) error {
		if upgradeHeight == ms.LastCommitID().Version+1 {
			// Check if the current commit version and upgrade height matches
			if len(storeUpgrades.Deleted) > 0 || len(storeUpgrades.Added) > 0 {
				stup := &storetypes.StoreUpgrades{
					Added:   storeUpgrades.Added,
					Deleted: storeUpgrades.Deleted,
				}
				return ms.LoadLatestVersionAndUpgrade(stup)
			}
		}

		// Otherwise load default store loader
		return baseapp.DefaultStoreLoader(ms)
	}
}
