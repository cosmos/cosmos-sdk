package simapp

import (
	"context"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v053 to v055, the default upgrade path for v0.55.x.
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.53.x to v0.55.x in a single coordinated upgrade.
const UpgradeName = "v053-to-v055"

// UpgradeNameFromV054 defines the on-chain upgrade name for the sample SimApp
// upgrade from v054 to v055.
const UpgradeNameFromV054 = "v054-to-v055"

// UpgradeNameFromV055 defines the on-chain upgrade name for the sample SimApp
// upgrade from v0.55.0 to v0.55.1, which runs the x/feegrant v2 to v3 store
// migration.
const UpgradeNameFromV055 = "v055-to-v0551"

// storeUpgrades lists the store changes for each sample upgrade. The stores
// deleted when coming from v0.53.x belong to modules that SimApp no longer
// mounts in v0.55.x.
var storeUpgrades = map[string]storetypes.StoreUpgrades{
	UpgradeName: {
		Deleted: []string{"circuit", "group", "nft", "protocolpool"},
	},
	UpgradeNameFromV054: {
		Deleted: []string{"protocolpool"},
	},
	UpgradeNameFromV055: {},
}

func (app SimApp) RegisterUpgradeHandlers() {
	for name := range storeUpgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			name,
			func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
				sdk.UnwrapSDKContext(ctx).Logger().Debug("this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
				return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
			},
		)
	}

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	upgrades, ok := storeUpgrades[upgradeInfo.Name]
	if ok && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrades))
	}
}
