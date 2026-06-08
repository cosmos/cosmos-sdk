package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type Upgrade[T AppI] struct {
	Name            string
	StoreUpgrades   storetypes.StoreUpgrades
	UpgradeCallBack func(ctx sdk.Context, plan upgradetypes.Plan, app T) error
}

func RegisterUpgradeHandlers[T AppI](app T, upgrades ...Upgrade[T]) {
	if app.UpgradeKeeper() == nil {
		panic("upgrade keeper is nil")
	}

	seen := make(map[string]struct{}, len(upgrades))
	for _, upgrade := range upgrades {
		if _, dup := seen[upgrade.Name]; dup {
			panic(fmt.Sprintf("duplicate upgrade name %q", upgrade.Name))
		}
		seen[upgrade.Name] = struct{}{}

		app.UpgradeKeeper().SetUpgradeHandler(
			upgrade.Name,
			func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
				sdkCtx := sdk.UnwrapSDKContext(ctx)
				sdkCtx.Logger().Debug("running upgrade handler", zap.String("upgrade_name", upgrade.Name))

				if upgrade.UpgradeCallBack != nil {
					err := upgrade.UpgradeCallBack(sdkCtx, plan, app)
					if err != nil {
						return nil, err
					}
				}

				return app.ModuleManager().RunMigrations(ctx, app.Configurator(), fromVM)
			})
	}

	// Read upgrade info once outside the loop — it is the same file for all
	// upgrades and reading it N times introduces a TOCTOU window.
	upgradeInfo, err := app.UpgradeKeeper().ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	for _, upgrade := range upgrades {
		if upgradeInfo.Name == upgrade.Name && !app.UpgradeKeeper().IsSkipHeight(upgradeInfo.Height) {
			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
			break
		}
	}
}
