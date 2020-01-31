package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const UpgradeInfoFileName string = "upgrade-info.json"

// UpgradeStoreLoader is used to prepare baseapp with a fixed StoreLoader
// pattern. This is useful in test cases, or with custom upgrade loading logic.
func UpgradeStoreLoader(storeUpgrades *storetypes.StoreUpgrades) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {

		// Check if the current commit version and upgrade height matches
		if len(storeUpgrades.Renamed) > 0 || len(storeUpgrades.Deleted) > 0 {
			var lastBlockHeight = ms.LastCommitID().Version
			upgradeInfo, err := ReadUpgradeInfoFromDisk()

			// There should not be any error in reading upgrade info from disk
			// when there are store-upgrades planned.
			if err != nil {
				return err
			}

			if upgradeInfo.Height == lastBlockHeight {
				return ms.LoadLatestVersionAndUpgrade(storeUpgrades)
			}
		}

		// Otherwise load default store loader
		return baseapp.DefaultStoreLoader(ms)
	}
}

func ReadUpgradeInfoFromDisk() (storetypes.UpgradeInfo, error) {
	var upgradeInfo storetypes.UpgradeInfo
	// TODO cleanup viper
	home := viper.GetString(flags.FlagHome)
	upgradeInfoPath := filepath.Join(home, UpgradeInfoFileName)

	_, err := os.Stat(upgradeInfoPath)
	if err != nil {
		return upgradeInfo, fmt.Errorf("upgrade-file is not found: %s", err.Error())
	}

	data, err := ioutil.ReadFile(upgradeInfoPath)
	if err != nil {
		return upgradeInfo, fmt.Errorf("error while reading upgrade-file from filesystem: %s", err.Error())
	}

	err = json.Unmarshal(data, &upgradeInfo)
	if err != nil {
		return upgradeInfo, fmt.Errorf("error while decoding upgrade-file from filesystem: %s", err.Error())
	}

	return upgradeInfo, err
}
