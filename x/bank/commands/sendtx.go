package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/builder"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	cryptokeys "github.com/tendermint/go-crypto/keys"
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
	// get the from address
	from, err := builder.GetFromAddress()
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

	// get account name
	name := viper.GetString(client.FlagName)

	// get password
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		return err
	}

	// build message
	msg := BuildMsg(from, to, coins)

	// build and sign the transaction, then broadcast to Tendermint
	res, err := builder.SignBuildBroadcast(name, passphrase, msg, c.Cdc)
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

func (c Commander) SignMessage(msg sdk.Msg, kb cryptokeys.Keybase, accountName string, password string) ([]byte, error) {
	// sign and build
	bz := msg.GetSignBytes()
	sig, pubkey, err := kb.Sign(accountName, password, bz)
	if err != nil {
		return nil, err
	}
	sigs := []sdk.StdSignature{{
		PubKey:    pubkey,
		Signature: sig,
		Sequence:  viper.GetInt64(client.FlagName),
	}}

	// marshal bytes
	tx := sdk.NewStdTx(msg, sigs)

	txBytes, err := c.Cdc.MarshalBinary(tx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}
