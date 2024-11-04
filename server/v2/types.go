package serverv2

import (
	"github.com/spf13/viper"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/store/v2"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	appmanager.StateTransitionFunction[T]

	Name() string
	InterfaceRegistry() server.InterfaceRegistry
	QueryHandlers() map[string]appmodulev2.Handler
	Store() store.RootStore
	SchemaDecoderResolver() decoding.DecoderResolver
	Close() error
}

type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, corestore.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (corestore.ReaderMap, error)
}
