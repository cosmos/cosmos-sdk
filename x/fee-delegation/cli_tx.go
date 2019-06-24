package fee_delegation

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	utils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
)

func GetCmdDelegateFees(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegate [grantee] [fee-allowance]",
		Short: "delegate",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			account := cliCtx.GetFromAddress()

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			var allowance FeeAllowance
			err = cdc.UnmarshalJSON([]byte(args[1]), &allowance)
			if err != nil {
				return err
			}

			msg := NewMsgDelegateFeeAllowance(account, grantee, allowance)

			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
}

func GetCmdRevokeDelegatedFees(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke [grantee]",
		Short: "revoke",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			account := cliCtx.GetFromAddress()

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := NewMsgRevokeFeeAllowance(account, grantee)

			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
}
