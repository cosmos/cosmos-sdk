package serverv2

import (
	"github.com/spf13/viper"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/server/v2/appmanager"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	Name() string
	InterfaceRegistry() server.InterfaceRegistry
	GetAppManager() *appmanager.AppManager[T]
	GetQueryHandlers() map[string]appmodulev2.Handler
	// It is used to get the module schema resolver for the indexer.
	GetSchemaDecoderResolver() decoding.DecoderResolver
}
