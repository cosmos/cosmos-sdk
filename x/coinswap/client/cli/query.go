package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/coinswap/internal/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	coinswapQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the coinswap module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	coinswapQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryLiquidity(queryRoute, cdc),
		GetCmdQueryParams(queryRoute, cdc))...)

	return coinswapQueryCmd
}

// GetCmdQueryLiquidity implements the liquidity query command
func GetCmdQueryLiquidity(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "liquidity",
		Short: "Query the current liquidity values",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the liquidity of a specific trading pair stored in the reserve pool.
			
Example:
$ %s query coinswap liquidity
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			bz, err := cdc.MarshalJSON(types.NewQueryLiquidityParams(args[0]))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryLiquidity)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var liquidity sdk.Coins
			cdc.MustUnmarshalJSON(res, &liquidity)
			return cliCtx.PrintOutput(liquidity)
		},
	}
}

// GetCmdQueryParams implements the params query command
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the parameters involved in the coinswap process",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all the parameters for the coinswap process.
			
Example:
$ %s query coinswap params
`,
				version.ClientName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryParameters)
			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			cdc.MustUnmarshalJSON(bz, &params)
			return cliCtx.PrintOutput(params)
		},
	}
}
