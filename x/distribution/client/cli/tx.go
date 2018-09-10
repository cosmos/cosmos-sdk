// nolint
package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	wire "github.com/tendermint/go-wire"
)

type TxWithdrawDelegationRewardsAll struct {
	delegatorAddr sdk.AccAddress
	withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

type TxWithdrawDelegationReward struct {
	delegatorAddr sdk.AccAddress
	validatorAddr sdk.AccAddress
	withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

type TxWithdrawValidatorRewardsAll struct {
	operatorAddr sdk.AccAddress // validator address to withdraw from
	withdrawAddr sdk.AccAddress // address to make the withdrawal to
}

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
)

// GetCmdDelegate implements the delegate command.
func GetCmdWithdrawDelegationRewardsAll(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [delegator]",
		Short: "withdraw rewards for all delegations",
		RunE: func(cmd *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}

			delAddr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			msg := distr.NewMsgDelegate(delAddr, valAddr, amount)

			// build and sign the transaction, then broadcast to Tendermint
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}

	// TODO add flags for "is-validator", "only-for-validator"
	cmd.Flags().String(flagOnlyFromValidator, "", "Only withdraw from this validator address")
	cmd.Flags().Bool(flagIsValidator, false, "Also withdraw validator's commission")

	return cmd
}
