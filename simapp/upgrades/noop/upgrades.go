package noop

import (
	"context"

	"cosmossdk.io/simapp/upgrades"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// NewUpgrade constructor
func NewUpgrade[T upgrades.AppKeepers](semver string) upgrades.Upgrade[T] {
	return upgrades.Upgrade[T]{
		UpgradeName:          semver,
		CreateUpgradeHandler: CreateUpgradeHandler[T],
		StoreUpgrades: storetypes.StoreUpgrades{
			Added:   []string{},
			Deleted: []string{},
		},
	}
}

func CreateUpgradeHandler[T upgrades.AppKeepers](
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	ak *T,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
