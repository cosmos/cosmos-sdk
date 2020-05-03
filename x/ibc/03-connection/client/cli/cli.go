package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the query commands for IBC connections
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	ics03ConnectionQueryCmd := &cobra.Command{
		Use:                        "connection",
		Short:                      "IBC connection query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	ics03ConnectionQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryConnections(queryRoute, cdc),
		GetCmdQueryConnection(queryRoute, cdc),
	)...)

	return ics03ConnectionQueryCmd
}

// GetTxCmd returns the transaction commands for IBC connections
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics03ConnectionTxCmd := &cobra.Command{
		Use:   "connection",
		Short: "IBC connection transaction subcommands",
	}

	ics03ConnectionTxCmd.AddCommand(flags.PostCommands(
		GetCmdConnectionOpenInit(storeKey, cdc),
		GetCmdConnectionOpenTry(storeKey, cdc),
		GetCmdConnectionOpenAck(storeKey, cdc),
		GetCmdConnectionOpenConfirm(storeKey, cdc),
	)...)

	return ics03ConnectionTxCmd
}
