package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
)

// SendTxCommand will create a send tx and sign it with the given key
func SendTxCmd(Cdc *wire.Codec) *cobra.Command {
	cmdr := Commander{Cdc}
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE:  cmdr.sendTxCmd,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	return cmd
}

type Commander struct {
	Cdc *wire.Codec
}

func (c Commander) sendTxCmd(cmd *cobra.Command, args []string) error {
	ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(c.Cdc))

	// get the from address
	from, err := ctx.GetFromAddress()
	if err != nil {
		return err
	}

	// parse coins
	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return err
	}

	// parse destination address
	dest := viper.GetString(flagTo)
	bz, err := hex.DecodeString(dest)
	if err != nil {
		return err
	}
	to := sdk.Address(bz)

	// build message
	msg := BuildMsg(from, to, coins)

	// default to next sequence number if none provided
	if viper.GetInt64(client.FlagSequence) == 0 {
		from, err := ctx.GetFromAddress()
		if err != nil {
			return err
		}
		seq, err := ctx.NextSequence(from)
		if err != nil {
			return err
		}
		fmt.Printf("Defaulting to next sequence number: %d\n", seq)
		ctx = ctx.WithSequence(seq)
	}

	// build and sign the transaction, then broadcast to Tendermint
	res, err := ctx.SignBuildBroadcast(ctx.FromAddressName, msg, c.Cdc)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}

func BuildMsg(from sdk.Address, to sdk.Address, coins sdk.Coins) sdk.Msg {
	input := bank.NewInput(from, coins)
	output := bank.NewOutput(to, coins)
	msg := bank.NewSendMsg([]bank.Input{input}, []bank.Output{output})
	return msg
}
