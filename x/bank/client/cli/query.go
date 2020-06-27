package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	flagDenom = "denom"
)

func GetQueryCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetBalancesCmd(clientCtx),
		GetCmdQueryTotalSupply(clientCtx),
	)

	return cmd
}

func GetBalancesCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query for account balances by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			queryClient := types.NewQueryClient(clientCtx.Init())

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			denom := viper.GetString(flagDenom)
			pageReq := &query.PageRequest{}
			if denom == "" {
				params := types.NewQueryAllBalancesRequest(addr, pageReq)
				res, err := queryClient.AllBalances(context.Background(), params)
				if err != nil {
					return err
				}
				return clientCtx.PrintOutput(res.Balances)

			}

			params := types.NewQueryBalanceRequest(addr, denom)
			res, err := queryClient.Balance(context.Background(), params)
			if err != nil {
				return err
			}
			return clientCtx.PrintOutput(res.Balance)
		},
	}

	cmd.Flags().String(flagDenom, "", "The specific balance denomination to query for")
	return flags.GetCommands(cmd)[0]
}

func GetCmdQueryTotalSupply(clientCtx client.Context) *cobra.Command {
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
			queryClient := types.NewQueryClient(clientCtx.Init())

			if len(args) == 0 {
				res, err := queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
				if err != nil {
					return err
				}
				return clientCtx.PrintOutput(res.Supply)
			}

			res, err := queryClient.SupplyOf(context.Background(), &types.QuerySupplyOfRequest{Denom: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintOutput(res.Amount)
		},
	}

	return flags.GetCommands(cmd)[0]
}
