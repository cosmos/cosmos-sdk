package simapp

import (
	"context"

	"cosmossdk.io/core/appmodule"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts"
	epochstypes "cosmossdk.io/x/epochs/types"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v0.50.x to v0.51.x
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.50.x to v0.51.x.
const UpgradeName = "v050-to-v051"

func (app SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM appmodule.VersionMap) (appmodule.VersionMap, error) {
			// sync accounts and auth module account number
			err := authkeeper.MigrateAccountNumberUnsafe(ctx, &app.AuthKeeper)
			if err != nil {
				return nil, err
			}

			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := corestore.StoreUpgrades{
			Added: []string{
				accounts.StoreKey,
				protocolpooltypes.StoreKey,
				epochstypes.StoreKey,
				countertypes.StoreKey, // This module is used for testing purposes only.
			},
			Deleted: []string{"crisis"}, // The SDK discontinued the crisis module in v0.52.0
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
