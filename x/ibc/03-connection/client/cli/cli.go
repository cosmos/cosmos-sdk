package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// GetQueryCmd returns the query commands for IBC connections
func GetQueryCmd(clientCtx client.Context) *cobra.Command {
	ics03ConnectionQueryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC connection query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics03ConnectionQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryConnections(clientCtx),
		GetCmdQueryConnection(clientCtx),
		GetCmdQueryClientConnections(clientCtx),
	)...)

	return ics03ConnectionQueryCmd
}

// NewTxCmd returns a CLI command handler for all x/ibc connection transaction commands.
func NewTxCmd(clientCtx client.Context) *cobra.Command {
	ics03ConnectionTxCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC connection transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics03ConnectionTxCmd.AddCommand(flags.PostCommands(
		NewConnectionOpenInitCmd(clientCtx),
		NewConnectionOpenTryCmd(clientCtx),
		NewConnectionOpenAckCmd(clientCtx),
		NewConnectionOpenConfirmCmd(clientCtx),
	)...)

	return ics03ConnectionTxCmd
}
