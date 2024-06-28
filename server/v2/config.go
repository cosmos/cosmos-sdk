package serverv2

import "github.com/spf13/cobra"

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
