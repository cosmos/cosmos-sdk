// nolint
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *amino.Codec) *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:   "dist",
		Short: "Distribution transactions subcommands",
	}

	distTxCmd.AddCommand(client.PostCommands(
		GetCmdWithdrawRewards(cdc),
		GetCmdSetWithdrawAddr(cdc),
	)...)

	return distTxCmd
}

// command to withdraw rewards
func GetCmdWithdrawRewards(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards",
		Short: "withdraw rewards for either a delegation or a validator",
		Long: strings.TrimSpace(`Withdraw rewards from either a delegation or a validator:


`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			onlyFromVal := viper.GetString(flagOnlyFromValidator)
			isVal := viper.GetBool(flagIsValidator)

			if onlyFromVal != "" && isVal {
				return fmt.Errorf("cannot use --%v, and --%v flags together",
					flagOnlyFromValidator, flagIsValidator)
			}

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			var msg sdk.Msg
			switch {
			case isVal:
				addr := cliCtx.GetFromAddress()
				valAddr := sdk.ValAddress(addr.Bytes())
				msg = types.NewMsgWithdrawValidatorCommission(valAddr)
			default:
				delAddr := cliCtx.GetFromAddress()
				valAddr, err := sdk.ValAddressFromBech32(onlyFromVal)
				if err != nil {
					return err
				}

				msg = types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
			}

			return utils.MessageOutput(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	cmd.Flags().String(flagOnlyFromValidator, "", "only withdraw from this validator address (in bech)")
	cmd.Flags().Bool(flagIsValidator, false, "also withdraw validator's commission")
	return cmd
}

// command to withdraw all rewards
func GetCmdWithdrawAllRewards(cdc *codec.Codec, queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-all-rewards [delegator-addr]",
		Short: "withdraw all delegations rewards for a delegator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			delAddr := cliCtx.GetFromAddress()
			msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, cdc, queryRoute, delAddr)
			if err != nil {
				return err
			}

			return utils.MessageOutput(cliCtx, txBldr, msgs, false)
		},
	}
	cmd.Flags().String(flagOnlyFromValidator, "", "only withdraw from this validator address (in bech)")
	cmd.Flags().Bool(flagIsValidator, false, "also withdraw validator's commission")
	return cmd
}

// command to replace a delegator's withdrawal address
func GetCmdSetWithdrawAddr(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Long: strings.TrimSpace(`Set the withdraw address for rewards assoicated with a delegator address:

$ gaiacli tx set-withdraw-addr cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			delAddr := cliCtx.GetFromAddress()
			withdrawAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)
			return utils.MessageOutput(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}
