package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/coin"
)

// SendTxCmd is CLI command to send tokens between basecoin accounts
var SendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send tokens from one account to another",
	RunE:  commands.RequireInit(sendTxCmd),
}

//nolint
const (
	FlagTo     = "to"
	FlagAmount = "amount"
	FlagFrom   = "from"
)

func init() {
	flags := SendTxCmd.Flags()
	flags.String(FlagTo, "", "Destination address for the bits")
	flags.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	flags.String(FlagFrom, "", "Address sending coins, if not first signer")
}

// sendTxCmd is an example of how to make a tx
func sendTxCmd(cmd *cobra.Command, args []string) error {
	tx, err := readSendTxFlags()
	if err != nil {
		return err
	}

	tx, err = txcmd.Middleware.Wrap(tx)
	if err != nil {
		return err
	}

	err = txcmd.SignTx(tx)
	if err != nil {
		return err
	}

	// otherwise, post it and display response
	bres, err := txcmd.PrepareOrPostTx(tx)
	if err != nil {
		return err
	}
	if bres == nil {
		return nil // successful prep, nothing left to do
	}
	return txcmd.OutputTx(bres) // print response of the post
}

func readSendTxFlags() (tx basecoin.Tx, err error) {
	// parse to address
	toAddr, err := commands.ParseAddress(viper.GetString(FlagTo))
	if err != nil {
		return tx, err
	}

	fromAddr, err := readFromAddr()
	if err != nil {
		return tx, err
	}

	amountCoins, err := coin.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return tx, err
	}

	// craft the inputs and outputs
	ins := []coin.TxInput{{
		Address: fromAddr,
		Coins:   amountCoins,
	}}
	outs := []coin.TxOutput{{
		Address: toAddr,
		Coins:   amountCoins,
	}}

	return coin.NewSendTx(ins, outs), nil
}

func readFromAddr() (basecoin.Actor, error) {
	from := viper.GetString(FlagFrom)
	if from == "" {
		return txcmd.GetSignerAct(), nil
	}
	return commands.ParseAddress(from)
}
