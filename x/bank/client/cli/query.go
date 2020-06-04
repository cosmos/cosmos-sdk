package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	flagDenom = "denom"
)

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------

// GetQueryCmd returns the parent querying command for the bank module.
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetBalancesCmd(cdc),
		GetCmdQueryTotalSupply(cdc),
	)

	return cmd
}

// GetAccountCmd returns a CLI command handler that facilitates querying for a
// single or all account balances by address.
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
func GetBalancesCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query for account balances by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc)

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var (
				params interface{}
				result interface{}
				route  string
			)

			denom := viper.GetString(flagDenom)
			if denom == "" {
				params = types.NewQueryAllBalancesRequest(addr)
				route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllBalances)
			} else {
				params = types.NewQueryBalanceRequest(addr, denom)
				route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance)
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			res, _, err := clientCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			if denom == "" {
				var balances sdk.Coins
				if err := cdc.UnmarshalJSON(res, &balances); err != nil {
					return err
				}

				result = balances
			} else {
				var balance sdk.Coin
				if err := cdc.UnmarshalJSON(res, &balance); err != nil {
					return err
				}

				result = balance
			}

			return clientCtx.PrintOutput(result)
		},
	}

	cmd.Flags().String(flagDenom, "", "The specific balance denomination to query for")

	return flags.GetCommands(cmd)[0]
}

// TODO: Remove once client-side Protobuf migration has been completed.
// ref: https://github.com/cosmos/cosmos-sdk/issues/5864
func GetCmdQueryTotalSupply(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
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
			clientCtx := client.NewContext().WithCodec(cdc)

			if len(args) == 0 {
				return queryTotalSupply(clientCtx, cdc)
			}

			return querySupplyOf(clientCtx, cdc, args[0])
		},
	}

	return flags.GetCommands(cmd)[0]
}
