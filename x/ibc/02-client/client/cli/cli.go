package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// GetQueryCmd returns the query commands for IBC clients
func GetQueryCmd(clientCtx client.Context) *cobra.Command {
	ics02ClientQueryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC client query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics02ClientQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryClientStates(clientCtx),
		GetCmdQueryClientState(clientCtx),
		GetCmdQueryConsensusState(clientCtx),
		GetCmdQueryHeader(clientCtx),
		GetCmdNodeConsensusState(clientCtx),
		GetCmdQueryPath(clientCtx),
	)...)
	return ics02ClientQueryCmd
}
