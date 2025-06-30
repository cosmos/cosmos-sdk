package simapp

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/cosmos/gogoproto/jsonpb"

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
const (
	UpgradeName       = "v053-to-v054"
	ManualUpgradeName = "manual1"
)

func (app SimApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeName,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdk.UnwrapSDKContext(ctx).Logger().Debug("this is a debug level message to test that verbose logging mode has properly been enabled during a chain upgrade")
			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)
	// we add another upgrade, to be performed manually which does some small state breakage
	app.UpgradeKeeper.SetUpgradeHandler(
		ManualUpgradeName,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// do some minimal state breaking update
			err := app.GovKeeper.Constitution.Set(ctx,
				fmt.Sprintf("we have expected upgrade %q and that's now our constitution", plan.Name))
			return fromVM, err
		},
	)

	// we check that we can read the upgrade info from disk, which is necessary for setting store key upgrades
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}
	if upgradeInfo.Name != "" {
		app.Logger().Info("read upgrade info from disk", "upgrade_info", upgradeInfo)
	}

	// this allows to test stateful manual upgrades with Cosmovisor
	if manualUpgradeVar, ok := os.LookupEnv("SIMAPP_MANUAL_UPGRADE"); ok {
		var manualUpgrade upgradetypes.Plan
		err := (&jsonpb.Unmarshaler{}).Unmarshal(bytes.NewBufferString(manualUpgradeVar), &manualUpgrade)
		if err != nil {
			panic("invalid SIMAPP_MANUAL_UPGRADE: " + err.Error())
		}
		err = app.UpgradeKeeper.SetManualUpgrade(&manualUpgrade)
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
