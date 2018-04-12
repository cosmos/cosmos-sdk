package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"

	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
)

// take the coolness quiz transaction
func QuizTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cool [answer]",
		Short: "What's cooler than being cool?",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || len(args[0]) == 0 {
				return errors.New("You must provide an answer")
			}

			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			msg := cool.NewQuizMsg(from, args[0])

			// get account name
			name := viper.GetString(client.FlagName)

			// default to next sequence number if none provided
			ctx, err = context.EnsureSequence(ctx)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.SignBuildBroadcast(name, msg, cdc)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 || len(args[0]) == 0 {
				return errors.New("You must provide an answer")
			}

			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// get account name
			name := viper.GetString(client.FlagName)

			// default to next sequence number if none provided
			ctx, err = context.EnsureSequence(ctx)
			if err != nil {
				return err
			}

			// create the message
			msg := cool.NewSetTrendMsg(from, args[0])

			// build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.SignBuildBroadcast(name, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
}
