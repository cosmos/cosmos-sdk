package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
)

// GetQueryCmd returns the query commands for IBC connections
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC connection query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	queryCmd.AddCommand(
		GetCmdQueryConnections(),
		GetCmdQueryConnection(),
		GetCmdQueryClientConnections(),
	)

	return queryCmd
}

// NewTxCmd returns a CLI command handler for all x/ibc connection transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC connection transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewConnectionOpenInitCmd(),
		NewConnectionOpenTryCmd(),
		NewConnectionOpenAckCmd(),
		NewConnectionOpenConfirmCmd(),
	)

	return txCmd
}
