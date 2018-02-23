package tx

import (
	"github.com/spf13/cobra"
)

// AddCommands adds a number of tx-query related subcommands
func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		txSearchCommand(),
		txCommand(),
	)
}
