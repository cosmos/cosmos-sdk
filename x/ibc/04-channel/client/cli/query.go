package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// TODO: get proofs
// const (
// 	FlagProve = "prove"
// )

// GetQueryCmd returns the query commands for IBC channels
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelQueryCmd := &cobra.Command{
		Use:                "channel",
		Short:              "IBC channel query subcommands",
		DisableFlagParsing: true,
	}

	ics04ChannelQueryCmd.AddCommand(cli.GetCommands(
		GetCmdQueryChannel(storeKey, cdc),
	)...)

	return ics04ChannelQueryCmd
}

// GetCmdQueryChannel defines the command to query a channel end
func GetCmdQueryChannel(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "end [port-id] [channel-id]",
		Short: "Query stored connection",
		Long: strings.TrimSpace(fmt.Sprintf(`Query stored connection end
		
Example:
$ %s query ibc channel end [port-id] [channel-id]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			portID := args[0]
			channelID := args[1]

			bz, err := cdc.MarshalJSON(types.NewQueryChannelParams(portID, channelID))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(types.ChannelPath(portID, channelID), bz)
			if err != nil {
				return err
			}

			var channel types.Channel
			if err := cdc.UnmarshalJSON(res, &channel); err != nil {
				return err
			}

			return cliCtx.PrintOutput(channel)
		},
	}

	// cmd.Flags().Bool(FlagProve, false, "(optional) show proofs for the query results")

	return cmd
}
