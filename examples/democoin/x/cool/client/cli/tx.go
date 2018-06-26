package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
)

// take the coolness quiz transaction
func QuizTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cool [answer]",
		Short: "What's cooler than being cool?",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			msg := cool.NewMsgQuiz(from, args[0])

			// get account name
			name := viper.GetString(client.FlagName)

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.EnsureSignBuildBroadcast(name, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}

// set a new cool trend transaction
func SetTrendTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "setcool [answer]",
		Short: "You're so cool, tell us what is cool!",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// get account name
			name := viper.GetString(client.FlagName)

			// create the message
			msg := cool.NewMsgSetTrend(from, args[0])

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.EnsureSignBuildBroadcast(name, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}
