package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/params/types"
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

	cmd.AddCommand(NewQuerySubspaceParamsCmd())

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
			clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			params := types.NewQuerySubspaceParams(args[0], args[1])
			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)

			bz, err := clientCtx.JSONMarshaler.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			bz, _, err = clientCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.SubspaceParamsResponse
			if err := clientCtx.JSONMarshaler.UnmarshalJSON(bz, &resp); err != nil {
				return err
			}

			return clientCtx.PrintOutput(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
