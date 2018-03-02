package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/bank"
)

const (
	flagTo       = "to"
	flagAmount   = "amount"
	flagFee      = "fee"
	flagSequence = "seq"
)

// SendTxCommand will create a send tx and sign it with the given key
func SendTxCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := commander{cdc}
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE:  cmdr.sendTxCmd,
	}
	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")
	cmd.Flags().String(flagFee, "", "Fee to pay along with transaction")
	cmd.Flags().Int64(flagSequence, 0, "Sequence number to sign the tx")
	return cmd
}

type commander struct {
	cdc *wire.Codec
}

func (c commander) sendTxCmd(cmd *cobra.Command, args []string) error {
	txBytes, err := c.buildTx()
	if err != nil {
		return err
	}

	res, err := client.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}

func (c commander) buildTx() ([]byte, error) {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	name := viper.GetString(client.FlagName)
	info, err := keybase.Get(name)
	if err != nil {
		return nil, errors.Errorf("No key for: %s", name)
	}
	from := info.PubKey.Address()

	msg, err := buildMsg(from)
	if err != nil {
		return nil, err
	}

	// sign and build
	bz := msg.GetSignBytes()
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		return nil, err
	}
	sig, pubkey, err := keybase.Sign(name, passphrase, bz)
	if err != nil {
		return nil, err
	}
	sigs := []sdk.StdSignature{{
		PubKey:    pubkey,
		Signature: sig,
		Sequence:  viper.GetInt64(flagSequence),
	}}

	// marshal bytes
	tx := sdk.NewStdTx(msg, sigs)

	txBytes, err := c.cdc.MarshalBinary(tx)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func buildMsg(from sdk.Address) (sdk.Msg, error) {

	// parse coins
	amount := viper.GetString(flagAmount)
	coins, err := sdk.ParseCoins(amount)
	if err != nil {
		return nil, err
	}

	// parse destination address
	dest := viper.GetString(flagTo)
	bz, err := hex.DecodeString(dest)
	if err != nil {
		return nil, err
	}
	to := sdk.Address(bz)

	input := bank.NewInput(from, coins)
	output := bank.NewOutput(to, coins)
	msg := bank.NewSendMsg([]bank.Input{input}, []bank.Output{output})
	return msg, nil
}
