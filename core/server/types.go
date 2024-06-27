package server

import (
	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/transaction"
)

type AppCreator[T transaction.Tx] func(log.Logger, AppOptions) AppI[T]

type AppOptions interface {
	Get(string) interface{}
}

type AppI[T transaction.Tx] interface {
	GetAppManager() any
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetStore() any
}
