package cli

import (
	"fmt"

	"github.com/spf13/cobra"

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
func NewTxCmd() *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distTxCmd.AddCommand(
<<<<<<< HEAD
		NewWithdrawRewardsCmd(valAc, ac),
		NewWithdrawAllRewardsCmd(valAc, ac),
		NewSetWithdrawAddrCmd(ac),
		NewFundCommunityPoolCmd(ac),
		NewDepositValidatorRewardsPoolCmd(valAc, ac),
=======
		NewWithdrawAllRewardsCmd(),
>>>>>>> 44934e3b6 (feat(x/distribution): add autocli options for tx (#17963))
	)

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
			delAddr, err := clientCtx.AddressCodec.BytesToString(clientCtx.GetFromAddress())
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
				_, err := clientCtx.ValidatorAddressCodec.StringToBytes(valAddr)
				if err != nil {
					return err
				}

				msg := types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
				msgs = append(msgs, msg)
			}

			chunkSize, _ := cmd.Flags().GetInt(FlagMaxMessagesPerTx)
			if chunkSize == 0 {
				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
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
<<<<<<< HEAD

// NewSetWithdrawAddrCmd returns a CLI command handler for creating a MsgSetWithdrawAddress transaction.
func NewSetWithdrawAddrCmd(ac address.Codec) *cobra.Command {
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
			delAddr := clientCtx.GetFromAddress()
			withdrawAddr, err := ac.StringToBytes(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewFundCommunityPoolCmd returns a CLI command handler for creating a MsgFundCommunityPool transaction.
func NewFundCommunityPoolCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-community-pool [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Funds the community pool with the specified amount",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Funds the community pool with the specified amount

Example:
$ %s tx distribution fund-community-pool 100uatom --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			depositorAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundCommunityPool(amount, depositorAddr)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDepositValidatorRewardsPoolCmd returns a CLI command handler for creating
// a MsgDepositValidatorRewardsPool transaction.
func NewDepositValidatorRewardsPoolCmd(valCodec, ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-validator-rewards-pool [val_addr] [amount]",
		Args:  cobra.ExactArgs(2),
		Short: "Fund the validator rewards pool with the specified amount",
		Example: fmt.Sprintf(
			"%s tx distribution fund-validator-rewards-pool cosmosvaloper1x20lytyf6zkcrv5edpkfkn8sz578qg5sqfyqnp 100uatom --from mykey",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			depositorAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			_, err = valCodec.StringToBytes(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositValidatorRewardsPool(depositorAddr, args[0], amount)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
=======
>>>>>>> 44934e3b6 (feat(x/distribution): add autocli options for tx (#17963))
