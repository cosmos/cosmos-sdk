package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/client/utils"
)

// GetCmdQueryNextSequence defines the command to query a next receive sequence
func GetCmdQueryNextSequence(cdc *codec.Codec, queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-recv [port-id] [channel-id]",
		Short: "Query a next receive sequence",
		Long: strings.TrimSpace(fmt.Sprintf(`Query an IBC channel end
		
Example:
$ %s query ibc channel next-recv [port-id] [channel-id]
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc channel next-recv [port-id] [channel-id]", version.ClientName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			portID := args[0]
			channelID := args[1]
			prove := viper.GetBool(flags.FlagProve)

			sequenceRes, err := utils.QueryNextSequenceRecv(cliCtx, portID, channelID, prove)
			if err != nil {
				return err
			}

			cliCtx = cliCtx.WithHeight(int64(sequenceRes.ProofHeight))
			return cliCtx.PrintOutput(sequenceRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")

	return cmd
}
