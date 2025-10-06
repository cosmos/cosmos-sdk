package app

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

type Module interface {
	module.AppModule

	// additional providers //  TODO optional?
	StoreKeysProvider
	NameProvider
	MaccPermsProvider
}

type StoreKeysProvider interface {
	StoreKeys() map[string]*storetypes.KVStoreKey
}

type NameProvider interface {
	Name() string
}

type MaccPermsProvider interface {
	MaccPerms() map[string][]string
}
