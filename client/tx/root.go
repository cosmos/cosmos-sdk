package tx

import (
	"errors"

	"github.com/spf13/cobra"
)

const (
	flagTags = "tag"
	flagAny  = "any"
)

// XXX: remove this when not needed
func todoNotImplemented(_ *cobra.Command, _ []string) error {
	return errors.New("TODO: Command not yet implemented")
}

// AddCommands adds a number of tx-query related subcommands
func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		txSearchCommand(),
		txCommand(),
	)
}

func txSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringSlice(flagTags, nil, "Tags that must match (may provide multiple)")
	cmd.Flags().Bool(flagAny, false, "Return transactions that match ANY tag, rather than ALL")
	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx <hash>",
		Short: "Matches this txhash over all committed blocks",
		RunE:  todoNotImplemented,
	}
	return cmd
}
