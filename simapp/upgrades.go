package simapp

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/nft"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeName defines the on-chain upgrade name for the sample simap upgrade from v045 to v046.
const UpgradeName = "v045-to-v046"

func (app SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(UpgradeName,
		func(ctx sdk.Context, plan upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
			// 1st-time running in-store migrations, using 1 as fromVersion to
			// avoid running InitGenesis.
			// If you wish to skip any module migrations (already migrated in old version),
			// you can use `modulename.AppModule{}.ConsensusVersion()` instead of `1` below. For example:
			// "auth":	auth.AppModule{}.ConsensusVersion(),
			fromVM := map[string]uint64{
				"auth":         1,
				"authz":        1,
				"bank":         1,
				"capability":   1,
				"crisis":       1,
				"distribution": 1,
				"evidence":     1,
				"feegrant":     1,
				"gov":          1,
				"mint":         1,
				"params":       1,
				"slashing":     1,
				"staking":      1,
				"upgrade":      1,
				"vesting":      1,
				"genutil":      1,
			}

			return app.mm.RunMigrations(ctx, app.configurator, fromVM)
		})

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{group.ModuleName, nft.ModuleName},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}
