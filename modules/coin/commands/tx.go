package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"

	"github.com/tendermint/basecoin"
	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/modules/coin"
)

// SendTxCmd is CLI command to send tokens between basecoin accounts
var SendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send tokens from one account to another",
	RunE:  commands.RequireInit(doSendTx),
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

// doSendTx is an example of how to make a tx
func doSendTx(cmd *cobra.Command, args []string) error {
	// load data from json or flags
	// var tx basecoin.Tx
	// found, err := txcmd.LoadJSON(&tx)
	// if err != nil {
	// 	return err
	// }

	tx, err := readSendTxFlags()
	if err != nil {
		return err
	}

	tx, err = bcmd.Middleware.Wrap(tx)
	if err != nil {
		return err
	}

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(tx.Unwrap())
	if err != nil {
		return err
	}
	if err = bcmd.ValidateResult(bres); err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readSendTxFlags() (tx basecoin.Tx, err error) {
	// parse to address
	toAddr, err := bcmd.ParseAddress(viper.GetString(FlagTo))
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
		return bcmd.GetSignerAct(), nil
	}
	return bcmd.ParseAddress(from)
}
