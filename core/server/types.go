package server

import (
	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/transaction"
)

type AppI[T transaction.Tx] interface {
	GetAppManager() any
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetStore() any
}
