// nolint
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	amino "github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			onlyFromVal := viper.GetString(flagOnlyFromValidator)
			isVal := viper.GetBool(flagIsValidator)

			if onlyFromVal != "" && isVal {
				return fmt.Errorf("cannot use --%v, and --%v flags together",
					flagOnlyFromValidator, flagIsValidator)
			}

			var msg sdk.Msg
			switch {
			case isVal:
				addr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}
				valAddr := sdk.ValAddress(addr.Bytes())
				msg = types.NewMsgWithdrawValidatorCommission(valAddr)
			default:
				delAddr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}

				valAddr, err := sdk.ValAddressFromBech32(onlyFromVal)
				if err != nil {
					return err
				}

				msg = types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
			}

			// build and sign the transaction, then broadcast to Tendermint
			return cliCtx.MessageOutput(msg)
		},
	}
	cmd.Flags().String(flagOnlyFromValidator, "", "only withdraw from this validator address (in bech)")
	cmd.Flags().Bool(flagIsValidator, false, "also withdraw validator's commission")
	return cmd
}

// GetCmdDelegate implements the delegate command.
func GetCmdSetWithdrawAddr(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			withdrawAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			return cliCtx.MessageOutput(types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr))
		},
	}
}
