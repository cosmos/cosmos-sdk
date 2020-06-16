// nolint
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
	flagCommission        = "commission"
	flagMaxMessagesPerTx  = "max-msgs"
)

const (
	MaxMessagesPerTxDefault = 5
)

// NewTxCmd returns a root CLI command handler for all x/distribution transaction commands.
func NewTxCmd(clientCtx client.Context) *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distTxCmd.AddCommand(flags.PostCommands(
		NewWithdrawRewardsCmd(clientCtx),
		NewWithdrawAllRewardsCmd(clientCtx),
		NewSetWithdrawAddrCmd(clientCtx),
		NewFundCommunityPoolCmd(clientCtx),
	)...)

	return distTxCmd
}

type newGenerateOrBroadcastFunc func(clientCtx client.Context, msgs ...sdk.Msg) error

func newSplitAndApply(
	newGenerateOrBroadcast newGenerateOrBroadcastFunc,
	clientCtx client.Context,
	msgs []sdk.Msg,
	chunkSize int,
) error {
	if chunkSize == 0 {
		return newGenerateOrBroadcast(clientCtx, msgs...)
	}

	// split messages into slices of length chunkSize
	totalMessages := len(msgs)
	for i := 0; i < len(msgs); i += chunkSize {

		sliceEnd := i + chunkSize
		if sliceEnd > totalMessages {
			sliceEnd = totalMessages
		}

		msgChunk := msgs[i:sliceEnd]
		if err := newGenerateOrBroadcast(clientCtx, msgChunk...); err != nil {
			return err
		}
	}

	return nil
}

func NewWithdrawRewardsCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.

Example:
$ %s tx distribution withdraw-rewards cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
$ %s tx distribution withdraw-rewards cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey --commission
`,
				version.ClientName, version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			delAddr := clientCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)}
			if viper.GetBool(flagCommission) {
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(valAddr))
			}

			for _, msg := range msgs {
				if err := msg.ValidateBasic(); err != nil {
					return err
				}
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msgs...)
		},
	}
	cmd.Flags().Bool(flagCommission, false, "also withdraw validator's commission")
	return cmd
}

func NewWithdrawAllRewardsCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-all-rewards",
		Short: "withdraw all delegations rewards for a delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw all rewards for a single delegator.

Example:
$ %s tx distribution withdraw-all-rewards --from mykey
`,
				version.ClientName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			delAddr := clientCtx.GetFromAddress()

			// The transaction cannot be generated offline since it requires a query
			// to get all the validators.
			if clientCtx.Offline {
				return fmt.Errorf("cannot generate tx in offline mode")
			}

			msgs, err := common.WithdrawAllDelegatorRewards(clientCtx, types.QuerierRoute, delAddr)
			if err != nil {
				return err
			}

			chunkSize := viper.GetInt(flagMaxMessagesPerTx)
			return newSplitAndApply(tx.GenerateOrBroadcastTx, clientCtx, msgs, chunkSize)
		},
	}

	cmd.Flags().Int(flagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
	return cmd
}

func NewSetWithdrawAddrCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set the withdraw address for rewards associated with a delegator address.

Example:
$ %s tx distribution set-withdraw-addr cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			delAddr := clientCtx.GetFromAddress()
			withdrawAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}
	return cmd
}

func NewFundCommunityPoolCmd(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-community-pool [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Funds the community pool with the specified amount",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Funds the community pool with the specified amount

Example:
$ %s tx distribution fund-community-pool 100uatom --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			depositorAddr := clientCtx.GetFromAddress()
			amount, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundCommunityPool(amount, depositorAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	return cmd
}

// GetCmdSubmitProposal implements the command to submit a community-pool-spend proposal
func GetCmdSubmitProposal(clientCtx client.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool-spend [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a community pool spend proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a community pool spend proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal community-pool-spend <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Community Pool Spend",
  "description": "Pay me some Atoms!",
  "recipient": "cosmos1s5afhd6gxevu37mkqcvvsj8qeylhn0rz46zdlq",
  "amount": "1000stake",
  "deposit": "1000stake"
}
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := clientCtx.InitWithInput(cmd.InOrStdin())

			proposal, err := ParseCommunityPoolSpendProposalJSON(clientCtx.JSONMarshaler, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			amount, err := sdk.ParseCoins(proposal.Amount)
			if err != nil {
				return err
			}
			content := types.NewCommunityPoolSpendProposal(proposal.Title, proposal.Description, proposal.Recipient, amount)

			deposit, err := sdk.ParseCoins(proposal.Deposit)
			if err != nil {
				return err
			}
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTx(clientCtx, msg)
		},
	}

	return cmd
}
