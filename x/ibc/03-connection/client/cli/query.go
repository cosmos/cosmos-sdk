package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
)

// GetCmdQueryConnections defines the command to query all the connection ends
// that this chain mantains.
func GetCmdQueryConnections(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connections",
		Short: "Query all available light clients",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all available connections

Example:
$ %s query ibc connection connections
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc connection connections", version.ClientName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx = clientCtx.Init()

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			connections, height, err := utils.QueryAllConnections(clientCtx, page, limit)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(connections)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of light clients to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of light clients to query for")

	return cmd
}

// GetCmdQueryConnection defines the command to query a connection end
func GetCmdQueryConnection(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "end [connection-id]",
		Short: "Query stored connection end",
		Long: strings.TrimSpace(fmt.Sprintf(`Query stored connection end
		
Example:
$ %s query ibc connection end [connection-id]
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc connection end [connection-id]", version.ClientName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

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

	return cmd
}

// GetCmdQueryAllClientConnections defines the command to query a all the client connection paths.
func GetCmdQueryAllClientConnections(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paths",
		Short: "Query all stored client connection paths",
		Long: strings.TrimSpace(fmt.Sprintf(`Query all stored client connection paths
		
Example:
$ %s query ibc connection paths
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc connection paths", version.ClientName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx = clientCtx.Init()

			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)

			connectionPaths, height, err := utils.QueryAllClientConnectionPaths(clientCtx, page, limit)
			if err != nil {
				return err
			}

			clientCtx = clientCtx.WithHeight(height)
			return clientCtx.PrintOutput(connectionPaths)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of light clients to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of light clients to query for")

	return cmd
}

// GetCmdQueryClientConnections defines the command to query a client connections
func GetCmdQueryClientConnections(clientCtx client.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "path [client-id]",
		Short: "Query stored client connection paths",
		Long: strings.TrimSpace(fmt.Sprintf(`Query stored client connection paths
		
Example:
$ %s query ibc connection path [client-id]
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc connection path [client-id]", version.ClientName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx = clientCtx.Init()

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
}
