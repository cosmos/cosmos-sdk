package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

// Transaction flags for the x/distribution module
var (
	FlagCommission       = "commission"
	FlagMaxMessagesPerTx = "max-msgs"
)

const (
	MaxMessagesPerTxDefault = 0
)

// NewTxCmd returns a root CLI command handler for all x/distribution transaction commands.
func NewTxCmd() *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distTxCmd.AddCommand(NewWithdrawAllRewardsCmd())

	return distTxCmd
}

// NewWithdrawAllRewardsCmd returns a CLI command handler for creating a MsgWithdrawDelegatorReward transaction.
// This command is more powerful than AutoCLI generated command as it allows sending batch of messages.
func NewWithdrawAllRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw-all-rewards",
		Short:   "Withdraw all delegations rewards for a delegator",
		Example: fmt.Sprintf("%s tx distribution withdraw-all-rewards --from mykey", version.AppName),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)}

			if commission, _ := cmd.Flags().GetBool(FlagCommission); commission {
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(valAddr))
			}

			for _, msg := range msgs {
				if err := msg.ValidateBasic(); err != nil {
					return err
				}
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	cmd.Flags().Bool(FlagCommission, false, "Withdraw the validator's commission in addition to the rewards")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewWithdrawAllRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-all-rewards",
		Short: "withdraw all delegations rewards for a delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw all rewards for a single delegator.
Note that if you use this command with --%[2]s=%[3]s or --%[2]s=%[4]s, the %[5]s flag will automatically be set to 0.

Example:
$ %[1]s tx distribution withdraw-all-rewards --from mykey
`,
				version.AppName, flags.FlagBroadcastMode, flags.BroadcastSync, flags.BroadcastAsync, FlagMaxMessagesPerTx,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// The transaction cannot be generated offline since it requires a query
			// to get all the validators.
			if clientCtx.Offline {
				return errors.New("cannot generate tx in offline mode")
			}

			queryClient := types.NewQueryClient(clientCtx)
			delValsRes, err := queryClient.DelegatorValidators(cmd.Context(), &types.QueryDelegatorValidatorsRequest{DelegatorAddress: delAddr})
			if err != nil {
				return err
			}

			validators := delValsRes.Validators
			// build multi-message transaction
			msgs := make([]sdk.Msg, 0, len(validators))
			for _, valAddr := range validators {
				_, err := clientCtx.ValidatorAddressCodec.StringToBytes(valAddr)
				if err != nil {
					return err
				}

				msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
				msgs = append(msgs, msg)
			}

			chunkSize, _ := cmd.Flags().GetInt(FlagMaxMessagesPerTx)
			if clientCtx.BroadcastMode != flags.BroadcastBlock && chunkSize > 0 {
				return fmt.Errorf("cannot use broadcast mode %[1]s with %[2]s != 0",
					clientCtx.BroadcastMode, FlagMaxMessagesPerTx)
			}

			return newSplitAndApply(tx.GenerateOrBroadcastTxCLI, clientCtx, cmd.Flags(), msgs, chunkSize)
		},
	}

	cmd.Flags().Int(FlagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewSetWithdrawAddrCmd() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set the withdraw address for rewards associated with a delegator address.

Example:
$ %s tx distribution set-withdraw-addr %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// split messages into slices of length chunkSize
			totalMessages := len(msgs)
			for i := 0; i < len(msgs); i += chunkSize {

				sliceEnd := i + chunkSize
				if sliceEnd > totalMessages {
					sliceEnd = totalMessages
				}

				msgChunk := msgs[i:sliceEnd]
				if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgChunk...); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().Int(FlagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
