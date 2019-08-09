package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/coinswap/internal/types"
)

// Liquidity flags
const (
	MinReward = "min-reward"
	MinNative = "min-native"
	Deadline  = "deadline"
	Recipient = "recipient"
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
		GetCmdRemoveLiquidity(cdc),
		GetCmdBuyOrder(cdc),
		GetCmdSellOrder(cdc))...)

	return coinswapTxCmd
}

// GetCmdAddLiquidity implements the add liquidity command handler
func GetCmdAddLiquidity(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [deposit-coin] [deposit]",
		Args:  cobra.ExactArgs(2),
		Short: "Add liquidity to the reserve pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add liquidity to the reserve pool for a trading pair.
			
Example:
$ %s tx coinswap add-liquidity dai 1000atom --min-reward 100 --deadline 2h --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			depositCoin, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			minRewardArg := viper.GetString(MinReward)
			minReward, err := sdk.ParseCoin(minRewardArg)
			if err != nil {
				return err
			}

			durationArg := viper.GetString(Deadline)
			duration, err := time.ParseDuration(durationArg)
			if err != nil {
				return fmt.Errorf("failed to parse the duration : %s", err)
			}
			deadline := time.Now().Add(duration).UTC()

			senderAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgAddLiquidity(depositCoin, deposit.Amount, minReward.Amount, deadline, senderAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(MinReward, "", "minimum amount of vouchers the sender is willing to accept for deposited coins (required)")
	cmd.Flags().String(Deadline, "1h", "duration for which the transaction is valid (required)")

	cmd.MarkFlagRequired(MinReward)
	cmd.MarkFlagRequired(Deadline)

	return cmd
}

// GetCmdRemoveLiquidity implements the remove liquidity command handler
func GetCmdRemoveLiquidity(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [withdraw-coin] [pool-tokens]",
		Args:  cobra.ExactArgs(2),
		Short: "Remove liquidity from the reserve pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove liquidity from the reserve pool for a trading pair.
			
Example:
$ %s tx coinswap remove-liquidity dai 1000atom --min-native 100atom --deadline 2h --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			withdrawCoin, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			poolTokens, ok := sdk.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("pool-tokens %s is not a valid int, please input valid pool-tokens", args[1])
			}

			minNativeArg := viper.GetString(MinNative)
			minNative, err := sdk.ParseCoin(minNativeArg)
			if err != nil {
				return err
			}

			durationArg := viper.GetString(Deadline)
			duration, err := time.ParseDuration(durationArg)
			if err != nil {
				return fmt.Errorf("failed to parse the duration : %s", err)
			}
			deadline := time.Now().Add(duration).UTC()

			senderAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgRemoveLiquidity(withdrawCoin, poolTokens, minNative.Amount, deadline, senderAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(MinNative, "", "minimum amount of the native asset the sender is willing to accept (required)")
	cmd.Flags().String(Deadline, "1h", "duration for which the transaction is valid (required)")

	cmd.MarkFlagRequired(MinNative)
	cmd.MarkFlagRequired(Deadline)

	return cmd
}

// GetCmdBuyOrder implements the buy order command handler
func GetCmdBuyOrder(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy-order [input] [output]",
		Args:  cobra.ExactArgs(2),
		Short: "Buy order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Buy order for a trading pair.
			
Example:
$ %s tx coinswap buy-order 5atom 2eth --deadline 2h --recipient recipientAddr --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			input, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			output, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			durationArg := viper.GetString(Deadline)
			duration, err := time.ParseDuration(durationArg)
			if err != nil {
				return fmt.Errorf("failed to parse the duration : %s", err)
			}
			deadline := time.Now().Add(duration).UTC()

			senderAddr := cliCtx.GetFromAddress()

			recipientAddrArg := viper.GetString(Recipient)
			recipientAddr, err := sdk.AccAddressFromBech32(recipientAddrArg)
			if err != nil {
				return err
			}

			msg := types.NewMsgSwapOrder(input, output, deadline, senderAddr, recipientAddr, true)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(Recipient, "", "recipient's address (required)")
	cmd.Flags().String(Deadline, "1h", "duration for which the transaction is valid (required)")

	cmd.MarkFlagRequired(Recipient)
	cmd.MarkFlagRequired(Deadline)

	return cmd
}

// GetCmdSellOrder implements the sell order command handler
func GetCmdSellOrder(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell-order [input] [output]",
		Args:  cobra.ExactArgs(2),
		Short: "Sell order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Sell order for a trading pair.
			
Example:
$ %s tx coinswap sell-order 2eth 5atom --deadline 2h --recipient recipientAddr --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			input, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			output, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			durationArg := viper.GetString(Deadline)
			duration, err := time.ParseDuration(durationArg)
			if err != nil {
				return fmt.Errorf("failed to parse the duration : %s", err)
			}
			deadline := time.Now().Add(duration).UTC()

			senderAddr := cliCtx.GetFromAddress()

			recipientAddrArg := viper.GetString(Recipient)
			recipientAddr, err := sdk.AccAddressFromBech32(recipientAddrArg)
			if err != nil {
				return err
			}

			msg := types.NewMsgSwapOrder(input, output, deadline, senderAddr, recipientAddr, false)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(Recipient, "", "recipient's address (required)")
	cmd.Flags().String(Deadline, "1h", "duration for which the transaction is valid (required)")

	cmd.MarkFlagRequired(Recipient)
	cmd.MarkFlagRequired(Deadline)

	return cmd
}
