package serverv2

import (
	"github.com/spf13/viper"

	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
)

type AppCreator[AppT AppI[T], T transaction.Tx] func(log.Logger, *viper.Viper) AppT

type AppI[T transaction.Tx] interface {
	GetAppManager() appmanager.AppManager[T]
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetStore() any
}
