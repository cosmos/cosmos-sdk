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

// StoreLoaderWithUpgrade is used to prepare baseapp with a fixed StoreLoader
// pattern. This is useful in test cases, or with custom upgrade loading logic.
func StoreLoaderWithUpgrade(storeUpgrades *storetypes.StoreUpgrades) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		// TODO cleanup viper
		home := viper.GetString(flags.FlagHome)
		upgradeInfoPath := filepath.Join(home, "upgrade-info.json")

		var lastBlockHeight = ms.LastCommitID().Version
		upgrades, err := GetUpgradeInfoDataFromFile(upgradeInfoPath)

		// There should not be any error in reading upgrade info from filesystem
		if err != nil {
			return err
		}

		// Check if the current commit version and upgrade height matches
		if (len(storeUpgrades.Renamed) > 0 || len(storeUpgrades.Deleted) > 0) &&
			upgrades.Height == lastBlockHeight {
			return ms.LoadLatestVersionAndUpgrade(storeUpgrades)
		}

		// Otherwise load default store loader
		return baseapp.DefaultStoreLoader(ms)
	}
}

func GetUpgradeInfoDataFromFile(upgradeInfoPath string) (storetypes.UpgradeInfo, error) {
	var upgrades storetypes.UpgradeInfo
	_, err := os.Stat(upgradeInfoPath)
	if err != nil {
		return upgrades, fmt.Errorf("upgrade-file is not found: %s", err.Error())
	}

	data, err := ioutil.ReadFile(upgradeInfoPath)
	if err != nil {
		return upgrades, fmt.Errorf("error while reading upgrade-file from filesystem: %s", err.Error())
	}

	err = json.Unmarshal(data, &upgrades)
	if err != nil {
		return upgrades, fmt.Errorf("error while decoding upgrade-file from filesystem: %s", err.Error())
	}

	return upgrades, err
}
