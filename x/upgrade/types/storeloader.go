package types

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2"
	"github.com/cosmos/cosmos-sdk/store/v2/multi"
)

// UpgradeStoreOption is used to prepare baseapp with a fixed StoreOption.
// This is useful for custom upgrade loading logic.
func UpgradeStoreOption(upgradeHeight uint64, storeUpgrades *storetypes.StoreUpgrades) baseapp.StoreOption {
	return func(par *multi.StoreParams, loadHeight uint64) error {
		// Check if the current commit version and upgrade height matches
		if upgradeHeight == loadHeight+1 {
			if len(storeUpgrades.Renamed) > 0 || len(storeUpgrades.Deleted) > 0 || len(storeUpgrades.Added) > 0 {
				par.Upgrades = storeUpgrades
			}
		}
		return nil
	}
}
