package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the query commands for IBC clients
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	ics02ClientQueryCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "IBC client query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics02ClientQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryClientStates(queryRoute, cdc),
		GetCmdQueryClientState(queryRoute, cdc),
		GetCmdQueryConsensusState(queryRoute, cdc),
		GetCmdQueryHeader(cdc),
		GetCmdNodeConsensusState(queryRoute, cdc),
		GetCmdQueryPath(queryRoute, cdc),
	)...)
	return ics02ClientQueryCmd
}
