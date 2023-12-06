package serverv2

import "github.com/spf13/cobra"

var ServerContextKey = struct{}{}

// Config is the config of the main server.
type Config struct {
	// StartBlock indicates if the server should block or not.
	// If true, the server will block until the context is canceled.
	StartBlock bool
}

// CLIConfig defines the CLI configuration for a module server.
type CLIConfig struct {
	// Command defines the main command of a module server.
	Command []*cobra.Command
	// Query defines the query commands of a module server.
	// Those commands are meant to be added in the root query command.
	Query []*cobra.Command
	// Tx defines the tx commands of a module server.
	// Those commands are meant to be added in the root tx command.
	Tx []*cobra.Command
}
