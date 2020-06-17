package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// GetQueryCmd returns the query commands for IBC channels
func GetQueryCmd(clientCtx client.Context) *cobra.Command {
	ics04ChannelQueryCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC channel query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics04ChannelQueryCmd.AddCommand(flags.GetCommands(
		// TODO: Query all channels
		GetCmdQueryChannel(clientCtx),
		// TODO: Query channels from a connection
		GetCmdQueryChannelClientState(clientCtx),
		// TODO: Query all packet commitments
		// TODO: Query unrelayed packet ACKS
		// TODO: Query unrelayed packet sends
	)...)

	return ics04ChannelQueryCmd
}

// NewTxCmd returns a CLI command handler for all x/ibc channel transaction commands.
func NewTxCmd(clientCtx client.Context) *cobra.Command {
	ics04ChannelTxCmd := &cobra.Command{
		Use:                        types.SubModuleName,
		Short:                      "IBC channel transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ics04ChannelTxCmd.AddCommand(flags.PostCommands(
		NewChannelOpenInitCmd(clientCtx),
		NewChannelOpenTryCmd(clientCtx),
		NewChannelOpenAckCmd(clientCtx),
		NewChannelOpenConfirmCmd(clientCtx),
		NewChannelCloseInitCmd(clientCtx),
		NewChannelCloseConfirmCmd(clientCtx),
	)...)

	return ics04ChannelTxCmd
}
