package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
)

// QuizTxCmd invokes the coolness quiz transaction.
func QuizTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cool [answer]",
		Short: "What's cooler than being cool?",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			msg := cool.NewMsgQuiz(from, args[0])

			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
}

// SetTrendTxCmd sends a new cool trend transaction.
func SetTrendTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "setcool [answer]",
		Short: "You're so cool, tell us what is cool!",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			msg := cool.NewMsgSetTrend(from, args[0])

			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
}
