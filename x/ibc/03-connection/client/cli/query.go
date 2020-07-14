package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// GetCmdQueryConnections defines the command to query all the connection ends
// that this chain mantains.
func GetCmdQueryConnections() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "connections",
		Short:   "Query all connections",
		Long:    "Query all connections ends from a chain",
		Example: fmt.Sprintf("%s query %s %s connections", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			offset, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			req := &types.QueryConnectionsRequest{
				Req: &query.PageRequest{
					Offset: uint64(offset),
					Limit:  uint64(limit),
				},
			}

			res, err := queryClient.Connections(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of light clients to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of light clients to query for")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryConnection defines the command to query a connection end
func GetCmdQueryConnection() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "end [connection-id]",
		Short:   "Query stored connection end",
		Long:    "Query stored connection end",
		Example: fmt.Sprintf("%s query %s %s end [connection-id]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			connectionID := args[0]
			prove, _ := cmd.Flags().GetBool(flags.FlagProve)

			connRes, err := utils.QueryConnection(clientCtx, connectionID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(connRes.ProofHeight))
			return clientCtx.PrintOutput(connRes)
		},
	}

	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryClientConnections defines the command to query a client connections
func GetCmdQueryClientConnections() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "path [client-id]",
		Short:   "Query stored client connection paths",
		Long:    "Query stored client connection paths",
		Example: fmt.Sprintf("%s query  %s %s path [client-id]", version.AppName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientID := args[0]
			prove, _ := cmd.Flags().GetBool(flags.FlagProve)

			connPathsRes, err := utils.QueryClientConnections(clientCtx, clientID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(connPathsRes.ProofHeight))
			return clientCtx.PrintOutput(connPathsRes)
		},
	}

	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
