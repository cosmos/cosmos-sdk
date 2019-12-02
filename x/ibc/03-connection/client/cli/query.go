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
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
)

// GetCmdQueryConnection defines the command to query a connection end
func GetCmdQueryConnection(queryRoute string, cdc *codec.Codec) *cobra.Command {
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
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			connectionID := args[0]
			prove := viper.GetBool(flags.FlagProve)

			connRes, err := utils.QueryConnection(cliCtx, connectionID, prove)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(connRes)
		},
	}
	cmd.Flags().Bool(flags.FlagProve, true, "show proofs for the query results")

	return cmd
}

// GetCmdQueryClientConnections defines the command to query a client connections
func GetCmdQueryClientConnections(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "client [client-id]",
		Short: "Query stored client connection paths",
		Long: strings.TrimSpace(fmt.Sprintf(`Query stored client connection paths
		
Example:
$ %s query ibc connection client [client-id]
		`, version.ClientName),
		),
		Example: fmt.Sprintf("%s query ibc connection client [client-id]", version.ClientName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			clientID := args[0]
			prove := viper.GetBool(flags.FlagProve)

			connPathsRes, err := utils.QueryClientConnections(cliCtx, clientID, prove)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(connPathsRes)
		},
	}
}
