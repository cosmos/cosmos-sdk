package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

// NewQueryCmd returns a root CLI command handler for all x/params query commands.
func NewQueryCmd(m codec.JSONMarshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the params module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewQuerySubspaceParamsCmd(m))

	return cmd
}

// NewQuerySubspaceParamsCmd returns a CLI command handler for querying subspace
// parameters managed by the x/params module.
func NewQuerySubspaceParamsCmd(m codec.JSONMarshaler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subspace [subspace] [key]",
		Short: "Query for raw parameters by subspace and key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithJSONMarshaler(m)

			params := types.NewQuerySubspaceParams(args[0], args[1])
			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)

			bz, err := m.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			bz, _, err = cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.SubspaceParamsResponse
			if err := m.UnmarshalJSON(bz, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}

	return flags.GetCommands(cmd)[0]
}
