package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
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
func NewTxCmd(valAc, ac address.Codec) *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distTxCmd.AddCommand(
		NewWithdrawRewardsCmd(valAc, ac),
		NewWithdrawAllRewardsCmd(valAc, ac),
	)

	return distTxCmd
}

type newGenerateOrBroadcastFunc func(client.Context, *pflag.FlagSet, ...sdk.Msg) error

func newSplitAndApply(
	genOrBroadcastFn newGenerateOrBroadcastFunc, clientCtx client.Context,
	fs *pflag.FlagSet, msgs []sdk.Msg, chunkSize int,
) error {
	if chunkSize == 0 {
		return genOrBroadcastFn(clientCtx, fs, msgs...)
	}

	// split messages into slices of length chunkSize
	totalMessages := len(msgs)
	for i := 0; i < len(msgs); i += chunkSize {

		sliceEnd := i + chunkSize
		if sliceEnd > totalMessages {
			sliceEnd = totalMessages
		}

		msgChunk := msgs[i:sliceEnd]
		if err := genOrBroadcastFn(clientCtx, fs, msgChunk...); err != nil {
			return err
		}
	}

	return nil
}

// NewWithdrawRewardsCmd returns a CLI command handler for creating a MsgWithdrawDelegatorReward transaction.
func NewWithdrawRewardsCmd(valCodec, ac address.Codec) *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.

Example:
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey --commission
`,
				version.AppName, bech32PrefixValAddr, version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{types.NewMsgWithdrawDelegatorReward(delAddr, args[0])}

			if commission, _ := cmd.Flags().GetBool(FlagCommission); commission {
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(args[0]))
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	cmd.Flags().Bool(FlagCommission, false, "Withdraw the validator's commission in addition to the rewards")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewWithdrawAllRewardsCmd returns a CLI command handler for creating a MsgWithdrawDelegatorReward transaction.
func NewWithdrawAllRewardsCmd(valCodec, ac address.Codec) *cobra.Command {
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
			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			// The transaction cannot be generated offline since it requires a query
			// to get all the validators.
			if clientCtx.Offline {
				return fmt.Errorf("cannot generate tx in offline mode")
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
				_, err := valCodec.StringToBytes(valAddr)
				if err != nil {
					return err
				}

				msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
				msgs = append(msgs, msg)
			}

			chunkSize, _ := cmd.Flags().GetInt(FlagMaxMessagesPerTx)

			return newSplitAndApply(tx.GenerateOrBroadcastTxCLI, clientCtx, cmd.Flags(), msgs, chunkSize)
		},
	}

	cmd.Flags().Int(FlagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
