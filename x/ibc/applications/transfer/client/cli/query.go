package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

// GetCmdQueryDenomTrace defines the command to query a a denomination trace from a given hash.
func GetCmdQueryDenomTrace() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "denom-trace [hash]",
		Short:   "Query the denom trace info from a given trace hash",
		Long:    "Query the denom trace info from a given trace hash",
		Example: fmt.Sprintf("%s query ibc-transfer denom-trace [hash]", version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
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

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryDenomTraces defines the command to query all the denomination trace infos
// that this chain mantains.
func GetCmdQueryDenomTraces() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "denom-traces",
		Short:   "Query the trace info for all token denominations",
		Long:    "Query the trace info for all token denominations",
		Example: fmt.Sprintf("%s query ibc-transfer denom-traces", version.AppName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
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

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "denominations trace")

	return cmd
}

// GetCmdParams returns the command handler for ibc-transfer parameter querying.
func GetCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "params",
		Short:   "Query the current ibc-transfer parameters",
		Long:    "Query the current ibc-transfer parameters",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf("%s query ibc-transfer params", version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, _ := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			return clientCtx.PrintProto(res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdParams returns the command handler for ibc-transfer parameter querying.
func GetCmdQueryEscrowAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "escrow-address",
		Short:   "Get the escrow address for a channel",
		Long:    "Get the escrow address for a channel",
		Args:    cobra.ExactArgs(2),
		Example: fmt.Sprintf("%s query ibc-transfer escrow-address [port] [channel-id]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			port := args[0]
			channel := args[1]
			addr := types.GetEscrowAddress(port, channel)
			fmt.Println(addr.String())
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
