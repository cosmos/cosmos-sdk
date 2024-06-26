package server

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	coreapp "cosmossdk.io/core/app"
	"cosmossdk.io/core/log"
	"cosmossdk.io/core/transaction"
)

type AppCreator[T transaction.Tx] func(log.Logger, *viper.Viper) AppI[T]

type AppI[T transaction.Tx] interface {
	GetAppManager() any
	GetConsensusAuthority() string
	InterfaceRegistry() coreapp.InterfaceRegistry
	GetStore() any
}

// CLIConfig defines the CLI configuration for a module server.
type CLIConfig struct {
	// Commands defines the main command of a module server.
	Commands []*cobra.Command
	// Queries defines the query commands of a module server.
	// Those commands are meant to be added in the root query command.
	Queries []*cobra.Command
	// Txs defines the tx commands of a module server.
	// Those commands are meant to be added in the root tx command.
	Txs []*cobra.Command
}