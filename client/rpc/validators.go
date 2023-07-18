package rpc

import (
	"strconv"

	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// TODO these next two functions feel kinda hacky based on their placement

// ValidatorCommand returns the validator set for a given height
func ValidatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tendermint-validator-set [height]",
		Short: "Get the full tendermint validator set at given height",
		Args:  cobra.MaximumNArgs(1),
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

			result, err := client.GetValidators(cmd.Context(), clientCtx, height, &page, &limit)
			if err != nil {
				return err
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "text", "Output format (text|json)")
	cmd.Flags().Int(flags.FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, 100, "Query number of results returned per page")

	return cmd
}
