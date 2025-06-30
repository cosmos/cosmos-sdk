package simapp

import (
	"context"
	"os"
	"strconv"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v053 to v054.
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.53.x to v0.54.x.
const UpgradeName = "v053-to-v054"

func (app SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdk.UnwrapSDKContext(ctx).Logger().Debug("this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}
	if upgradeInfo.Name != "" {
		app.Logger().Info("read upgrade info from disk", "upgrade_info", upgradeInfo)
	}

	// this allows us to check migration to v0.54.x in the system tests via a manual (non-governance upgrade)
	if manualUpgrade, ok := os.LookupEnv("SIMAPP_MANUAL_UPGRADE_HEIGHT"); ok {
		height, err := strconv.ParseUint(manualUpgrade, 10, 64)
		if err != nil {
			panic("invalid SIMAPP_MANUAL_UPGRADE_HEIGHT height: " + err.Error())
		}
		upgradeInfo = upgradetypes.Plan{
			Name:   UpgradeName,
			Height: int64(height),
		}
		err = app.UpgradeKeeper.SetManualUpgrade(&upgradeInfo)
		if err != nil {
			panic("failed to set manual upgrade: " + err.Error())
		}
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
