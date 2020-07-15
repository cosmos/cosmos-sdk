package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// NewQueryCmd returns a root CLI command handler for all x/params query commands.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the params module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewQuerySubspaceParamsCmd(),
	)

	return cmd
}

// NewQuerySubspaceParamsCmd returns a CLI command handler for querying subspace
// parameters managed by the x/params module.
func NewQuerySubspaceParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subspace [subspace] [key]",
		Short: "Query for raw parameters by subspace and key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := proposal.NewQueryClient(clientCtx)

			params := proposal.NewQueryParametersRequest(args[0], args[1])

			res, err := queryClient.Parameters(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.GetParams())
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
