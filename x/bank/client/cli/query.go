package cli

import (
	gocontext "context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	flagDenom = "denom"
)

// NewQueryCmd returns a root CLI command handler for all x/bank query commands.
func NewQueryCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewBalancesCmd(clientCtx),
		NewCmdQueryTotalSupply(clientCtx),
	)

	return cmd
}

// NewBalancesCmd returns a CLI command handler for querying account balance(s).
func NewBalancesCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query for account balance(s) by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx.QueryConn())
			denom := viper.GetString(flagDenom)
			if denom == "" {
				request := types.NewQueryAllBalancesRequest(addr)
				result, err := queryClient.AllBalances(gocontext.Background(), request)
				if err != nil {
					return err
				}
				return clientCtx.Println(result.Balances)
			} else {
				params := types.NewQueryBalanceRequest(addr, denom)
				result, err := queryClient.Balance(gocontext.Background(), params)
				if err != nil {
					return err
				}
				return clientCtx.Println(result)
			}
		},
	}

	cmd.Flags().String(flagDenom, "", "The specific balance denomination to query for")

	return flags.GetCommands(cmd)[0]
}

func NewCmdQueryTotalSupply(clientCtx client.Context) *cobra.Command {
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
			queryClient := types.NewQueryClient(clientCtx.QueryConn())

			if len(args) == 0 {
				res, err := queryClient.TotalSupply(gocontext.Background(), &types.TotalSupplyRequest{})
				if err != nil {
					return err
				}
				return clientCtx.PrintOutput(res.Balances)
			}

			res, err := queryClient.SupplyOf(gocontext.Background(), &types.SupplyOfRequest{Denom: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintOutput(res.Amount)
		},
	}

	return flags.GetCommands(cmd)[0]
}
