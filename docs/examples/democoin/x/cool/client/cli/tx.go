package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/docs/examples/democoin/x/cool"
)

// QuizTxCmd invokes the coolness quiz transaction.
func QuizTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cool [answer]",
		Short: "What's cooler than being cool?",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			return cliCtx.MessageOutput(cool.NewMsgQuiz(from, args[0]))
		},
	}
}

// SetTrendTxCmd sends a new cool trend transaction.
func SetTrendTxCmd(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "setcool [answer]",
		Short: "You're so cool, tell us what is cool!",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			return cliCtx.MessageOutput(cool.NewMsgSetTrend(from, args[0]))
		},
	}
}
