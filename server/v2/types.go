package serverv2

import (
	"github.com/spf13/viper"

	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
)

type AppCreator[T transaction.Tx] func(*viper.Viper, log.Logger) AppI[T]

type AppI[T transaction.Tx] interface {
	GetAppManager() *appmanager.AppManager[T]
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetStore() any
}
