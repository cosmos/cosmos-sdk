package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/coinswap/internal/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	coinswapTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Coinswap transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	coinswapTxCmd.AddCommand(client.PostCommands(
		GetCmdAddLiquidity(cdc),
		GetCmdRemoveLiquidity(cdc))...)

	return coinswapTxCmd
}

// GetCmdAddLiquidity implements the add liquidity command handler
func GetCmdAddLiquidity(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "add-liquidity [deposit] [deposit-amount] [min-reward] [deadline]",
		Args:  cobra.ExactArgs(4),
		Short: "Add liquidity in the reserve pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add liquidity in the reserve pool for a trading pair.
			
Example:
$ %s tx coinswap add-liquidity dai 1000atom 1000 2 cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			deposit, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			depositAmount, ok := sdk.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf(types.ErrNotPositive(types.DefaultCodespace, "deposit amount provided is not positive").Error())
			}

			minReward, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf(types.ErrNotPositive(types.DefaultCodespace, "minimum liquidity is not positive").Error())
			}

			deadline, err := time.Parse(time.RFC3339, args[3])
			if err != nil {
				return err
			}

			senderAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgAddLiquidity(deposit, depositAmount, minReward, deadline, senderAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdRemoveLiquidity implements the remove liquidity command handler
func GetCmdRemoveLiquidity(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-liquidity [withdraw] [withdraw-amount] [min-native] [deadline]",
		Args:  cobra.ExactArgs(4),
		Short: "Remove liquidity from the reserve pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove liquidity from the reserve pool for a trading pair.
			
Example:
$ %s tx coinswap remove-liquidity dai 1000atom 1000 2 cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			withdraw, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			withdrawAmount, ok := sdk.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf(types.ErrNotPositive(types.DefaultCodespace, "withdraw amount provided is not positive").Error())
			}

			minNative, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf(types.ErrNotPositive(types.DefaultCodespace, "minimum native amount is not positive").Error())
			}

			deadline, err := time.Parse(time.RFC3339, args[3])
			if err != nil {
				return err
			}

			senderAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgRemoveLiquidity(withdraw, withdrawAmount, minNative, deadline, senderAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
