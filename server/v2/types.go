package serverv2

import (
	"github.com/spf13/viper"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/store/v2"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	appmanager.AppManager[T]

	Name() string
	InterfaceRegistry() server.InterfaceRegistry
	QueryHandlers() map[string]appmodulev2.Handler
	Store() store.RootStore
	SchemaDecoderResolver() decoding.DecoderResolver
	Close() error
}
