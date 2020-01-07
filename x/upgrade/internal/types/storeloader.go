package types

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"io/ioutil"
	"os"
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
		if (len(upgrades.StoreUpgrades.Renamed) > 0 || len(upgrades.StoreUpgrades.Deleted) > 0) &&
			upgrades.Height == lastBlockHeight {
			return ms.LoadLatestVersionAndUpgrade(storeUpgrades)
		}

		// Otherwise load default store loader
		return baseapp.DefaultStoreLoader(ms)
	}
}

// UpgradeableStoreLoader can be configured by SetStoreLoader() to check for the
// existence of a given upgrade file - json encoded StoreUpgrades data.
//
// If not file is present, it will perform the default load (no upgrades to store).
//
// If the file is present, it will parse the file and execute those upgrades
// (rename or delete stores), while loading the data. It will also delete the
// upgrade file upon successful load, so that the upgrade is only applied once,
// and not re-applied on next restart
//
// This is useful for in place migrations when a store key is renamed between
// two versions of the software.
func UpgradeableStoreLoader(upgradeInfoPath string) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		var lastBlockHeight = ms.LastCommitID().Version
		upgrades, err := GetUpgradeInfoDataFromFile(upgradeInfoPath)

		// There should not be any error in reading upgrade info from filesystem
		// Binary should panic
		if err != nil {
			return err
		}

		// If the current upgrade has StoreUpgrades planned and the binary is loading for the first time
		// i.e., upgrades.Height == currentHeight
		// then do LoadLatestVersionAndUpgrade
		// Else, do DefaultStoreLoader
		if (len(upgrades.StoreUpgrades.Renamed) > 0 || len(upgrades.StoreUpgrades.Deleted) > 0) &&
			upgrades.Height == lastBlockHeight {
			err := ms.LoadLatestVersionAndUpgrade(&upgrades.StoreUpgrades)
			if err != nil {
				return fmt.Errorf("load and upgrade database: %v", err)
			}

			// if we have a successful load, we set the values to default
			upgrades.StoreUpgrades = storetypes.StoreUpgrades{
				Renamed: []storetypes.StoreRename{{
					OldKey: "",
					NewKey: "",
				}},
				Deleted: []string{""},
			}

			writeInfo, _ := json.Marshal(upgrades)

			// Write the successful upgrade information to file, so it doesn't check on loading the binary next time
			// We don't care if there's any error in updating the upgrade-info.json file
			// as the height changes and it doesn't effect any further after successful upgrade
			err = ioutil.WriteFile(upgradeInfoPath, writeInfo, 0644)

			// There should not be any error in writing the upgrade info to file.
			// Otherwise it will lead to restart the multistore upgrades every time when the binary restarts.
			if err != nil {
				return fmt.Errorf("error in multistore upgrade: %v", err)
			}

			return nil
		} else {
			return baseapp.DefaultStoreLoader(ms)
		}
	}
}

func GetUpgradeInfoDataFromFile(upgradeInfoPath string) (storetypes.UpgradeInfo, error) {
	var upgrades storetypes.UpgradeInfo
	_, err := os.Stat(upgradeInfoPath)

	// If the upgrade-info file is not found, panic with error
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
