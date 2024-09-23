package cli

import (
	"errors"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"

	"cosmossdk.io/x/bank/v2/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagDenom = "denom"
)

// GetQueryCmd returns the parent command for all x/bank CLi query commands. The
// provided clientCtx should have, at a minimum, a verifier, Tendermint RPC client,
// and marshaler set.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetBalanceCmd(),
	)

	return cmd
}

func GetBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [address] [denom]",
		Short: "Query an account balance by address and denom",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			denom := args[1]
			if denom == "" {
				return errors.New("empty denom")
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			req := types.NewQueryBalanceRequest(addr.String(), denom)
			out := new(types.QueryBalanceResponse)

			err = clientCtx.Invoke(ctx, gogoproto.MessageName(&types.QueryBalanceRequest{}), req, out)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(out)
		},
	}

	cmd.Flags().String(FlagDenom, "", "The specific balance denomination to query for")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all balances")

	return cmd
}
