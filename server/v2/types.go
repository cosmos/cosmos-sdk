package serverv2

import (
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/appmanager"
	coreapp "cosmossdk.io/core/app"
)

type Application[T transaction.Tx] interface {
	GetAppManager() *appmanager.AppManager[T]
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	// GetStore() any
}
