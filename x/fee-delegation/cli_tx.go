package fee_delegation

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/spf13/cobra"
)

func GetCmdDelegateFees(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegate [grantee] [fee-allowance]",
		Short: "delegate",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

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

			cliCtx.PrintResponse = true

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
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			account := cliCtx.GetFromAddress()

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := NewMsgRevokeFeeAllowance(account, grantee)

			cliCtx.PrintResponse = true

			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
}
