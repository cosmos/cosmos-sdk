package serverv2

import (
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"github.com/spf13/viper"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	Name() string
	InterfaceRegistry() server.InterfaceRegistry
	GetAppManager() *appmanager.AppManager[T]
	GetQueryHandlers() map[string]appmodulev2.Handler
	GetStore() any
}
