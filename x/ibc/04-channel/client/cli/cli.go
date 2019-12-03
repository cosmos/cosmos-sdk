package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetQueryCmd returns the query commands for IBC channels
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelQueryCmd := &cobra.Command{
		Use:                "channel",
		Short:              "IBC channel query subcommands",
		DisableFlagParsing: true,
	}

	ics04ChannelQueryCmd.AddCommand(client.GetCommands(
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

	ics04ChannelTxCmd.AddCommand(client.PostCommands(
		GetMsgChannelOpenInitCmd(storeKey, cdc),
		GetMsgChannelOpenTryCmd(storeKey, cdc),
		GetMsgChannelOpenAckCmd(storeKey, cdc),
		GetMsgChannelOpenConfirmCmd(storeKey, cdc),
		GetMsgChannelCloseInitCmd(storeKey, cdc),
		GetMsgChannelCloseConfirmCmd(storeKey, cdc),
		GetCmdHandshakeChannel(storeKey, cdc),
	)...)

	return ics04ChannelTxCmd
}
