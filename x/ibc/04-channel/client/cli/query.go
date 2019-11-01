package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	cli "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// GetQueryCmd returns the query commands for IBC channels
func GetQueryCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
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
func GetCmdQueryChannel(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "end [port-id] [channel-id]",
		Short: "Query a channel end",
		Long: strings.TrimSpace(fmt.Sprintf(`Query an IBC channel end
		
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

			req := abci.RequestQuery{
				Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.SubModuleName),
				Data:  bz,
				Prove: viper.GetBool(flags.FlagProve),
			}

			res, err := cliCtx.QueryABCI(req)
			if err != nil {
				return err
			}

			var channel types.Channel
			if err := cdc.UnmarshalJSON(res.Value, &channel); err != nil {
				return err
			}

			if res.Proof == nil {
				return cliCtx.PrintOutput(channel)
			}

			channelRes := types.NewChannelResponse(portID, channelID, channel, res.Proof, res.Height)
			return cliCtx.PrintOutput(channelRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")

	return cmd
}

func queryChannel(ctx client.CLIContext, portID string, channelID string, queryRoute string) (types.ChannelResponse, error) {
	var connRes types.ChannelResponse

	req := abci.RequestQuery{
		Path:  "store/ibc/key",
		Data:  types.KeyChannel(portID, channelID),
		Prove: true,
	}

	res, err := ctx.QueryABCI(req)
	if res.Value == nil || err != nil {
		return connRes, err
	}

	var channel types.Channel
	if err := ctx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &channel); err != nil {
		return connRes, err
	}
	return types.NewChannelResponse(portID, channelID, channel, res.Proof, res.Height), nil
}
