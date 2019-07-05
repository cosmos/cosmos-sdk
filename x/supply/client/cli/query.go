package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	// Group supply queries under a subcommand
	supplyQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the supply module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	supplyQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryTotalSupply(cdc),
	)...)

	return supplyQueryCmd
}

// GetCmdQueryTotalSupply implements the query total supply command.
func GetCmdQueryTotalSupply(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "total [denom]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Query the total supply of coins of the chain",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query total supply of coins that are held by accounts in the
			chain.

Example:
$ %s query %s total

To query for the total supply of a specific coin denomination use:
$ %s query %s total stake
`,
				version.ClientName, types.ModuleName, version.ClientName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			if len(args) == 0 {
				return queryTotalSupply(cliCtx, cdc)
			}
			return querySupplyOf(cliCtx, cdc, args[0])
		},
	}
}

func queryTotalSupply(cliCtx context.CLIContext, cdc *codec.Codec) error {
	params := types.NewQueryTotalSupplyParams(1, 0) // no pagination
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return err
	}

	res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryTotalSupply), bz)
	if err != nil {
		return err
	}

	var totalSupply sdk.Coins
	err = cdc.UnmarshalJSON(res, &totalSupply)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(totalSupply)
}

func querySupplyOf(cliCtx context.CLIContext, cdc *codec.Codec, denom string) error {
	params := types.NewQuerySupplyOfParams(denom)
	bz, err := cdc.MarshalJSON(params)
	if err != nil {
		return err
	}

	res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySupplyOf), bz)
	if err != nil {
		return err
	}

	var supply sdk.Int
	err = cdc.UnmarshalJSON(res, &supply)
	if err != nil {
		return err
	}

	return cliCtx.PrintOutput(supply)
}
