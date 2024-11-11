package serverv2

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"github.com/spf13/pflag"
)

// ServerComponent is a server module that can be started and stopped.
type ServerComponent[T transaction.Tx] interface {
	Name() string

	Start(context.Context) error
	Stop(context.Context) error
}

// HasStartFlags is a server module that has start flags.
type HasStartFlags interface {
	// StartCmdFlags returns server start flags.
	// Those flags should be prefixed with the server name.
	// They are then merged with the server config in one viper instance.
	StartCmdFlags() *pflag.FlagSet
}

// HasConfig is a server module that has a config.
type HasConfig interface {
	Config() any
}

// ConfigWriter is a server module that can write its config to a file.
type ConfigWriter interface {
	WriteConfig(path string) error
}

// HasCLICommands is a server module that has CLI commands.
type HasCLICommands interface {
	CLICommands() CLIConfig
}

// Store is a store interface that can be used by a server module.
type Store interface {
	// StateLatest returns a readonly view over the latest
	// committed state of the store. Alongside the version
	// associated with it.
	StateLatest() (uint64, store.ReaderMap, error)

	// StateAt returns a readonly view over the provided
	// state. Must error when the version does not exist.
	StateAt(version uint64) (store.ReaderMap, error)
}
