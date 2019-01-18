package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankClient "github.com/cosmos/cosmos-sdk/x/bank/client"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
)

// SendTxCmd will create a send tx and sign it with the given key.
func SendTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContextTx(cdc)

			addr, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			if err = cliCtx.EnsureAccountExists(addr); err != nil {
				return err
			}

			to, err := sdk.AccAddressFromBech32(viper.GetString(flagTo))
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoins(viper.GetString(flagAmount))
			if err != nil {
				return err
			}

			account, err := cliCtx.FetchAccount(addr)
			if err != nil {
				return err
			}

			// ensure account has enough coins
			if !account.GetCoins().IsAllGTE(coins) {
				return errors.Errorf("Address %s doesn't have enough coins to pay for this transaction.", addr)
			}

			// build and sign the transaction, then broadcast to Tendermint
			return cliCtx.MessageOutput(bankClient.CreateMsg(addr, to, coins))
		},
	}

	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.MarkFlagRequired(flagTo)
	cmd.MarkFlagRequired(flagAmount)

	return client.PostCommands(cmd)[0]
}
