package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
)

// GetTxCmd returns the transaction commands for IBC fungible token transfer
func GetQueryCmd(cdc *codec.Codec, storeKey string) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "transfer",
		Short: "IBC fungible token transfer query subcommands",
	}

	queryCmd.AddCommand(
		GetCmdQueryPacketProof(cdc, storeKey),
	)

	return queryCmd
}

// GetCmdQueryChannel defines the command to query a channel end
func GetCmdQueryPacketProof(cdc *codec.Codec, queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packet [port-id] [channel-id] [sequence]",
		Short: "Query a channel end",
		Long: strings.TrimSpace(fmt.Sprintf(`Query an IBC channel end
		
Example:
$ %s query ibc channel end [port-id] [channel-id]
		`, version.ClientName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// cliCtx := context.NewCLIContext().WithCodec(cdc)
			// portID := args[0]
			// channelID := args[1]

			// bz, err := cdc.MarshalJSON(types.New(portID, channelID))
			// if err != nil {
			// 	return err
			// }

			// req := abci.RequestQuery{
			// 	Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.SubModuleName),
			// 	Data:  bz,
			// 	Prove: viper.GetBool(flags.FlagProve),
			// }

			// res, err := cliCtx.QueryABCI(req)
			// if err != nil {
			// 	return err
			// }

			// var channel types.Channel
			// if err := cdc.UnmarshalJSON(res.Value, &channel); err != nil {
			// 	return err
			// }

			// if res.Proof == nil {
			// 	return cliCtx.PrintOutput(channel)
			// }

			// channelRes := types.NewChannelResponse(portID, channelID, channel, res.Proof, res.Height)
			// return cliCtx.PrintOutput(channelRes)
			return nil
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")

	return cmd
}
