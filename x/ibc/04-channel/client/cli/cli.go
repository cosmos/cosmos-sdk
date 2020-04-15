package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the query commands for IBC channels
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelQueryCmd := &cobra.Command{
		Use:                "channel",
		Short:              "IBC channel query subcommands",
		DisableFlagParsing: true,
	}

	ics04ChannelQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryChannel(storeKey, cdc),
	)...)

	return ics04ChannelQueryCmd
}

// GetTxCmd returns the transaction commands for IBC channels
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelTxCmd := &cobra.Command{
		Use:   "channel",
		Short: "IBC channel transaction subcommands",
	}

	ics04ChannelTxCmd.AddCommand(flags.PostCommands(
		GetMsgChannelOpenInitCmd(storeKey, cdc),
		GetMsgChannelOpenTryCmd(storeKey, cdc),
		GetMsgChannelOpenAckCmd(storeKey, cdc),
		GetMsgChannelOpenConfirmCmd(storeKey, cdc),
		GetMsgChannelCloseInitCmd(storeKey, cdc),
		GetMsgChannelCloseConfirmCmd(storeKey, cdc),
	)...)

	return ics04ChannelTxCmd
}
