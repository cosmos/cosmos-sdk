// nolint
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
)

// command to withdraw rewards
func GetCmdWithdrawRewards(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards",
		Short: "withdraw rewards for either: all-delegations, a delegation, or a validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			onlyFromVal := viper.GetString(flagOnlyFromValidator)
			isVal := viper.GetBool(flagIsValidator)

			if onlyFromVal != "" && isVal {
				return fmt.Errorf("cannot use --%v, and --%v flags together",
					flagOnlyFromValidator, flagIsValidator)
			}

			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			var msg sdk.Msg
			switch {
			case isVal:
				addr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}
				valAddr := sdk.ValAddress(addr.Bytes())
				msg = types.NewMsgWithdrawValidatorRewardsAll(valAddr)
			case onlyFromVal != "":
				delAddr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}

				valAddr, err := sdk.ValAddressFromBech32(onlyFromVal)
				if err != nil {
					return err
				}

				msg = types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
			default:
				delAddr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}
				msg = types.NewMsgWithdrawDelegatorRewardsAll(delAddr)
			}

			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagOnlyFromValidator, "", "only withdraw from this validator address (in bech)")
	cmd.Flags().Bool(flagIsValidator, false, "also withdraw validator's commission")
	return cmd
}

// GetCmdDelegate implements the delegate command.
func GetCmdSetWithdrawAddr(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			txBldr := authtxb.NewTxBuilderFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			withdrawAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)

			// build and sign the transaction, then broadcast to Tendermint
			return utils.CompleteAndBroadcastTxCli(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
	return cmd
}
