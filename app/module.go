package app

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

type Module interface {
	module.AppModule

	// additional providers //  TODO optional?
	Keeper
	NameProvider
	ModuleAccountPermissionsProvider
}

type Keeper interface {
	StoreKeysProvider
}

type StoreKeysProvider interface {
	StoreKeys() map[string]*storetypes.KVStoreKey
}

type NameProvider interface {
	Name() string
}

type ModuleAccountPermissionsProvider interface {
	ModuleAccountPermissions() map[string][]string
}
