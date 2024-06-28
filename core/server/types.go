package server

import (
	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/transaction"
)

type AppOptions interface {
	Get(string) interface{}
}

type AppCreator[AppT AppI[T], T transaction.Tx] func(log.Logger, AppOptions) AppT

type AppI[T transaction.Tx] interface {
	GetAppManager() any
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetStore() any
}
