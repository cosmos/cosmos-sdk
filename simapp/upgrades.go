package simapp

import (
	"context"

	"go.uber.org/zap"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/app"
	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
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

type Upgrade struct {
	Name            string
	StoreUpgrades   storetypes.StoreUpgrades
	UpgradeCallBack func(ctx sdk.Context, plan upgradetypes.Plan, app *app.SDKApp) error
}

var MyUpgrade = Upgrade{
	Name: UpgradeName,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			countertypes.ModuleName,
		},
	},
	UpgradeCallBack: func(ctx sdk.Context, plan upgradetypes.Plan, app *app.SDKApp) error {
		return nil
	},
}

func (app *SimApp) RegisterUpgradeHandlers(upgrades ...Upgrade) {
	if app.UpgradeKeeper == nil {
		panic("upgrade keeper is nil")
	}

	for _, upgrade := range upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.Name,
			func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				sdkCtx.Logger().Debug("running upgrade handler", zap.String("upgrade_name", upgrade.Name))

				err := upgrade.UpgradeCallBack(sdkCtx, plan, app.SDKApp)
				if err != nil {
					return nil, err
				}

				return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
			})

		upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
		if err != nil {
			panic(err)
		}

		if upgradeInfo.Name == upgrade.Name && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
		}
	}
}
