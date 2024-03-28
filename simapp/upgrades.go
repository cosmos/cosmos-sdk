package simapp

import (
	"fmt"

	"cosmossdk.io/simapp/upgrades"
	"cosmossdk.io/simapp/upgrades/noop"
	v051 "cosmossdk.io/simapp/upgrades/v051"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// Upgrades list of chain upgrades
var Upgrades = []upgrades.Upgrade[upgrades.AppKeepers]{v051.Upgrade}

// RegisterUpgradeHandlers registers the chain upgrade handlers
func (app SimApp) RegisterUpgradeHandlers() {
	if len(Upgrades) == 0 {
		// always have a unique upgrade registered for the current version to test in system tests or manual
		Upgrades = append(Upgrades, noop.NewUpgrade[upgrades.AppKeepers](app.Version()))
	}

	var keepers upgrades.AppKeepers
	app.GetStoreKeys()
	// register all upgrade handlers
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.ModuleManager,
				app.Configurator(),
				&keepers,
			),
		)
	}
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// register store loader for current upgrade
	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades)) // nolint:gosec
			break
		}
	}
}
