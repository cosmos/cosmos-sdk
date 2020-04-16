package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

// GetQueryCmd returns the query commands for IBC connections
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	ics03ConnectionQueryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC connection query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics03ConnectionQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryConnections(queryRoute, cdc),
		GetCmdQueryConnection(queryRoute, cdc),
	)...)
	return ics03ConnectionQueryCmd
}

// NewTxCmd returns a root CLI command handler for all x/ibc connection transaction commands.
func NewTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	ics03ConnectionTxCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC connection transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics03ConnectionTxCmd.AddCommand(flags.PostCommands(
		NewConnectionOpenInitCmd(m, txg, ar),
		NewConnectionOpenTryCmd(m, txg, ar),
		NewConnectionOpenAckCmd(m, txg, ar),
		NewConnectionOpenConfirmCmd(m, txg, ar),
	)...)
	return ics03ConnectionTxCmd
}
