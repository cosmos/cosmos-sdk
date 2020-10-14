package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// GetQueryCmd returns the query commands for IBC connections
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "ibc-convo",
		Short:                      "IBC conversation query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	queryCmd.AddCommand(
		GetCmdQueryPendingMessage(),
		GetCmdQueryInboxMessage(),
		GetCmdQueryOutboxMessage(),
	)

	return queryCmd
}

// NewTxCmd returns the transaction commands for IBC conversation
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "ibc-convo",
		Short:                      "IBC conversation transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewConvoTxCmd(),
	)

	return txCmd
}
