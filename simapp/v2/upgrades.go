package simapp

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/runtime/v2"
<<<<<<< HEAD
	"cosmossdk.io/x/accounts"
	epochstypes "cosmossdk.io/x/epochs/types"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
=======
	bankv2types "cosmossdk.io/x/bank/v2/types"
>>>>>>> 5581225a9 (fix(x/upgrade): register missing implementation for SoftwareUpgradeProposal (#23179))
	upgradetypes "cosmossdk.io/x/upgrade/types"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// UpgradeName defines the on-chain upgrade name for the sample SimApp upgrade
// from v0.50.x to v2
//
// NOTE: This upgrade defines a reference implementation of what an upgrade
// could look like when an application is migrating from Cosmos SDK version
// v0.50.x to v2.
const UpgradeName = "v050-to-v2"

func (app *SimApp[T]) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM appmodule.VersionMap) (appmodule.VersionMap, error) {
			if err := authkeeper.MigrateAccountNumberUnsafe(ctx, &app.AuthKeeper); err != nil {
				return nil, err
			}
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
<<<<<<< HEAD
				accounts.ModuleName,
				epochstypes.StoreKey,
				protocolpooltypes.ModuleName,
=======
				bankv2types.ModuleName,
>>>>>>> 5581225a9 (fix(x/upgrade): register missing implementation for SoftwareUpgradeProposal (#23179))
			},
		}

		app.SetStoreLoader(runtime.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
