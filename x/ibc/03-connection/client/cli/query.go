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
func GetCmdQueryConnections(clientCtx client.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "connections",
		Short:   "Query all connections",
		Long:    "Query all connections ends from a chain",
		Example: fmt.Sprintf("%s query %s %s connections", version.ClientName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			clientCtx = clientCtx.Init()
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryConnectionsRequest{
				Req: &query.PageRequest{},
			}

			res, err := queryClient.Connections(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res)
		},
	}
}

// GetCmdQueryConnection defines the command to query a connection end
func GetCmdQueryConnection(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "end [connection-id]",
		Short:   "Query stored connection end",
		Long:    "Query stored connection end",
		Example: fmt.Sprintf("%s query %s %s end [connection-id]", version.ClientName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

			connectionID := args[0]
			prove, err := cmd.Flags().GetBool(flags.FlagProve)
			if err != nil {
				return err
			}

			connRes, err := utils.QueryConnection(clientCtx, connectionID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(connRes.ProofHeight))
			return clientCtx.PrintOutput(connRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")

	return cmd
}

// GetCmdQueryAllClientConnections defines the command to query a all the client connection paths.
func GetCmdQueryAllClientConnections(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "paths",
		Short:   "Query all stored client connection paths",
		Long:    "Query all stored client connection paths",
		Example: fmt.Sprintf("%s query %s %s paths", version.ClientName, host.ModuleName, types.SubModuleName),
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			clientCtx = clientCtx.Init()
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryClientsConnectionsRequest{
				Req: &query.PageRequest{},
			}

			res, err := queryClient.ClientsConnections(context.Background(), req)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(res.Height)
			return clientCtx.PrintOutput(res)
		},
	}

	return cmd
}

// GetCmdQueryClientConnections defines the command to query a client connections
func GetCmdQueryClientConnections(clientCtx client.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "path [client-id]",
		Short:   "Query stored client connection paths",
		Long:    "Query stored client connection paths",
		Example: fmt.Sprintf("%s query  %s %s path [client-id]", version.ClientName, host.ModuleName, types.SubModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

			clientID := args[0]
			prove, err := cmd.Flags().GetBool(flags.FlagProve)
			if err != nil {
				return err
			}

			connPathsRes, err := utils.QueryClientConnections(clientCtx, clientID, prove)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(int64(connPathsRes.ProofHeight))
			return clientCtx.PrintOutput(connPathsRes)
		},
	}
}
