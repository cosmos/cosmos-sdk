package v051

import (
	"context"

	"cosmossdk.io/simapp/upgrades"
	"cosmossdk.io/x/accounts"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// UpgradeName defines the on-chain upgrade name
const UpgradeName = "v0.51"

var Upgrade = upgrades.Upgrade[upgrades.AppKeepers]{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			accounts.StoreKey,
			protocolpooltypes.StoreKey,
		},
		Deleted: []string{
			crisistypes.StoreKey, // The SDK discontinued the crisis module in v0.51.0
		},
	},
}

func CreateUpgradeHandler(
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	_ *upgrades.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
