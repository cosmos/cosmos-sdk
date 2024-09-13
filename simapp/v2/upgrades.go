package simapp

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/x/accounts"
	bankv2types "cosmossdk.io/x/bank/v2/types"
	epochstypes "cosmossdk.io/x/epochs/types"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v0.50.x to v0.51.x
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.50.x to v0.51.x.
const UpgradeName = "v050-to-v051"

func (app *SimApp[T]) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM appmodule.VersionMap) (appmodule.VersionMap, error) {
			return app.ModuleManager().RunMigrations(ctx, fromVM)
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{
			Added: []string{
				accounts.StoreKey,
				protocolpooltypes.StoreKey,
				epochstypes.StoreKey,
				bankv2types.ModuleName,
			},
			Deleted: []string{"crisis"}, // The SDK discontinued the crisis module in v0.52.0
		}

		app.SetStoreLoader(runtime.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
