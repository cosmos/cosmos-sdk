package upgrades

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// AppKeepers placeholder type to customize for concrete use cases
type AppKeepers any

type ModuleManager interface {
	RunMigrations(ctx context.Context, cfg module.Configurator, fromVM module.VersionMap) (module.VersionMap, error)
	GetVersionMap() module.VersionMap
}

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade[T AppKeepers] struct {
	// Upgrade version name, for the upgrade handler, e.g. `v0.51.0`
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	CreateUpgradeHandler func(ModuleManager, module.Configurator, *T) upgradetypes.UpgradeHandler
	StoreUpgrades        storetypes.StoreUpgrades
}
