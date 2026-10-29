package app

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Module interface {
	module.AppModule
	StoreKeysProvider
	NameProvider
	ModuleAccountPermissionsProvider
}

type StoreKeysProvider interface {
	StoreKeys() map[string]*storetypes.KVStoreKey
}

type TransientStoreKeysProvider interface {
	TransientStoreKeys() map[string]*storetypes.TransientStoreKey
}

type NameProvider interface {
	Name() string
}

type ModuleAccountPermissionsProvider interface {
	ModuleAccountPermissions() map[string][]string
}

// StakingHooksProvider is an optional Module interface for modules that need
// to register staking lifecycle hooks.
type StakingHooksProvider interface {
	StakingHooks() stakingtypes.StakingHooks
}
