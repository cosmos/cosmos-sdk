package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// NewTxCmd returns a root CLI command handler for all ibc localhost transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "Localhost client transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	txCmd.AddCommand(
		NewCreateClientCmd(),
	)

	return txCmd
}
