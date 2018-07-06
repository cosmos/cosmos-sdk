package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
	flagAsync  = "async"
)

// SendTxCmd will create a send tx and sign it with the given key
func SendTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from/to address
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			fromAcc, err := ctx.QueryStore(auth.AddressStoreKey(from), ctx.AccountStore)
			if err != nil {
				return err
			}

			bech32From := sdk.MustBech32ifyAcc(from)
			// Check if account was found
			if fromAcc == nil {
				return errors.New("No account with address " + bech32From +
					" was found in the state.\nAre you sure there has been a transaction involving it?")
			}

			toStr := viper.GetString(flagTo)

			to, err := sdk.GetAccAddressBech32(toStr)
			if err != nil {
				return err
			}
			// parse coins trying to be sent
			amount := viper.GetString(flagAmount)
			coins, err := sdk.ParseCoins(amount)
			if err != nil {
				return err
			}

			// ensure account has enough coins
			account, err := ctx.Decoder(fromAcc)
			if err != nil {
				return err
			}
			if !account.GetCoins().IsGTE(coins) {
				return errors.New("Address " + bech32From +
					" doesn't have enough coins to pay for this transaction.")
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := client.BuildMsg(from, to, coins)

			if viper.GetBool(flagAsync) {
				res, err := ctx.EnsureSignBuildBroadcastAsync(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
				if err != nil {
					return err
				}
				fmt.Println("Async tx sent. tx hash: ", res.Hash.String())
				return nil
			}
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil

		},
	}

	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().Bool(flagAsync, false, "Pass the async flag to send a tx without waiting for the tx to be included in a block")

	return cmd
}
