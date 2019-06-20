package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group supply queries under a subcommand
	supplyQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the supply module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	supplyQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryTotalSupply(queryRoute, cdc),
		GetCmdQuerySupplyOf(queryRoute, cdc))...)

	return supplyQueryCmd
}

// GetCmdQueryTotalSupply implements the query total supply command.
func GetCmdQueryTotalSupply(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total",
		Args:  cobra.ExactArgs(0),
		Short: "Query the total supply of coins of the chain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query total supply of coins that are held by accounts in the
			chain.

Example:
$ %s query %s total
`,
				version.ClientName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryTotalSupply))
			if err != nil {
				return err
			}

			var totalSupply sdk.Coins
			err := cdc.UnmarshalJSON(res, &totalSupply)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(totalSupply)
		},
	}
}

// GetCmdQuerySupplyOf implements the query supply of a coin command.
func GetCmdQuerySupplyOf(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total-of [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the total supply of a specific coin denomination",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the total supply of a specific coin denomination that is held by accounts in the
			chain.

Example:
$ %s query %s total-of stake
`,
				version.ClientName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			params := types.NewQuerySupplyOfParams(args[0])

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QuerySupplyOf), bz)
			if err != nil {
				return err
			}

			var totalSupply sdk.Coins
			err := cdc.UnmarshalJSON(res, &totalSupply)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(totalSupply)
		},
	}
}
