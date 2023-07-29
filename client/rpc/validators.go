package rpc

import (
	"strconv"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// ValidatorCommand returns the validator set for a given height
func ValidatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "comet-validator-set [height]",
		Aliases: []string{"cometbft-validator-set", "tendermint-validator-set"},
		Short:   "Get the full CometBFT validator set at given height",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			var height *int64

			// optional height
			if len(args) > 0 {
				val, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return err
				}

				if val > 0 {
					height = &val
				}
			}

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			node, err := clientCtx.GetNode()
			if err != nil {
				return err
			}

			validatorsRes, err := node.Validators(cmd.Context(), height, &page, &limit)
			if err != nil {
				return err
			}

			output, err := cmtjson.Marshal(validatorsRes)
			if err != nil {
				return err
			}

			return clientCtx.PrintRaw(output)
		},
	}

	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to CometBFT RPC interface for this chain")
	cmd.Flags().StringP(flags.FlagOutput, "o", "text", "Output format (text|json)")
	cmd.Flags().Int(flags.FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, 100, "Query number of results returned per page")

	return cmd
}
