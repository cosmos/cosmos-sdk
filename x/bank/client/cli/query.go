package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	flagDenom = "denom"
)

// GetQueryCmd returns the parent querying command for the bank module.
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(GetBalancesCmd(cdc))

	return cmd
}

// GetAccountCmd returns a CLI command handler that facilitates querying for a
// single or all account balances by address.
func GetBalancesCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query for account balances by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

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
				params = types.NewQueryAllBalancesParams(addr)
				route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAllBalances)
			} else {
				params = types.NewQueryBalanceParams(addr, denom)
				route = fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryBalance)
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			res, _, err := cliCtx.QueryWithData(route, bz)
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

			return cliCtx.PrintOutput(result)
		},
	}

	cmd.Flags().String(flagDenom, "", "The specific balance denomination to query for")

	return flags.GetCommands(cmd)[0]
}
