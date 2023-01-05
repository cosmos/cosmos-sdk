package cmd

import (
	"github.com/spf13/cobra"
)

// ConfigComamnd contains all the confix commands
// These command can be used to interactively update an application config value.
func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Utilities for managing application configuration",
	}

	// add subcommands
	cmd.AddCommand(
		MigrateCommand(),
		GetCommand(),
		SetCommand(),
		DiffCommand(),
	)

	return cmd
}
