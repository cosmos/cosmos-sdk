package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// GetQueryCmd returns the query commands for IBC channels
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelQueryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC channel query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics04ChannelQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryChannel(storeKey, cdc),
	)...)

	return ics04ChannelQueryCmd
}

// NewTxCmd returns a root CLI command handler for all x/ibc channel transaction commands.
func NewTxCmd(m codec.Marshaler, txg tx.Generator, ar tx.AccountRetriever) *cobra.Command {
	ics04ChannelTxCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC channel transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics04ChannelTxCmd.AddCommand(flags.PostCommands(
		NewChannelOpenInitTxCmd(m, txg, ar),
		NewChannelOpenTryTxCmd(m, txg, ar),
		NewChannelOpenAckTxCmd(m, txg, ar),
		NewChannelOpenConfirmTxCmd(m, txg, ar),
		NewChannelCloseInitTxCmd(m, txg, ar),
		NewChannelCloseConfirmTxCmd(m, txg, ar),
	)...)

	return ics04ChannelTxCmd
}
