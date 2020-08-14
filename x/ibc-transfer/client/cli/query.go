package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

// GetCmdQueryDenomTraces defines the command to query all the denomination trace information
// that this chain mantains.
func GetCmdQueryDenomTraces() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "denom-traces",
		Short:   "Query the trace info for all token denominations",
		Long:    "Query the trace info for all token denominations",
		Example: fmt.Sprintf("%s query ibc-transfer denom-traces", version.AppName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &types.QueryDenomTracesRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.DenomTraces(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "denomination traces")

	return cmd
}

// GetCmdQueryDenomTrace defines the command to query a denomination trace from a given hash.
func GetCmdQueryDenomTrace() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "denom-trace [hash]",
		Short:   "Query the denom trace info from a given trace hash",
		Long:    "Query the denom trace info from a given trace hash",
		Example: fmt.Sprintf("%s query ibc-transfer denom-trace [hash]", version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDenomTraceRequest{
				Hash: args[0],
			}

			res, err := queryClient.DenomTrace(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
