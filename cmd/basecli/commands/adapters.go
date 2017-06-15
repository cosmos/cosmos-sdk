package commands

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"

	btypes "github.com/tendermint/basecoin/types"
)

/*** Here is the sendtx command ***/

var SendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send tokens from one account to another",
	RunE:  doSendTx,
}

const (
	ToFlag       = "to"
	AmountFlag   = "amount"
	FeeFlag      = "fee"
	GasFlag      = "gas"
	SequenceFlag = "sequence"
)

func init() {
	flags := SendTxCmd.Flags()
	flags.String(ToFlag, "", "Destination address for the bits")
	flags.String(AmountFlag, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	flags.String(FeeFlag, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	flags.Int64(GasFlag, 0, "Amount of gas for this transaction")
	flags.Int(SequenceFlag, -1, "Sequence number for this transaction")
}

// runDemo is an example of how to make a tx
func doSendTx(cmd *cobra.Command, args []string) error {
	tx := new(btypes.SendTx)

	// load data from json or flags
	found, err := txcmd.LoadJSON(tx)
	if !found {
		err = readSendTxFlags(tx)
	}
	if err != nil {
		return err
	}

	send := &SendTx{
		chainID: viper.GetString(commands.ChainFlag),
		Tx:      tx,
	}
	send.AddSigner(txcmd.GetSigner())

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(send)
	if err != nil {
		return err
	}

	// output result
	return txcmd.OutputTx(bres)
}

func readSendTxFlags(tx *btypes.SendTx) error {
	// parse to address
	to, err := hex.DecodeString(StripHex(viper.GetString(ToFlag)))
	if err != nil {
		return errors.Errorf("To address is invalid hex: %v\n", err)
	}

	//parse the fee and amounts into coin types
	tx.Fee, err = btypes.ParseCoin(viper.GetString(FeeFlag))
	if err != nil {
		return err
	}
	amountCoins, err := btypes.ParseCoins(viper.GetString(AmountFlag))
	if err != nil {
		return err
	}

	// set the gas
	tx.Gas = viper.GetInt64(GasFlag)

	// craft the inputs and outputs
	tx.Inputs = []btypes.TxInput{{
		Coins:    amountCoins,
		Sequence: viper.GetInt(SequenceFlag),
	}}
	tx.Outputs = []btypes.TxOutput{{
		Address: to,
		Coins:   amountCoins,
	}}

	return nil
}

/******** AppTx *********/

type AppFlags struct {
	Fee      string
	Gas      int64
	Amount   string
	Sequence int
}

func AppFlagSet() (*flag.FlagSet, AppFlags) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String("amount", "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.String("fee", "", "Coins for the transaction fee of the format <amt><coin>")
	fs.Int64("gas", 0, "Amount of gas for this transaction")
	fs.Int("sequence", -1, "Sequence number for this transaction")
	return fs, AppFlags{}
}

// AppTxReader allows us to create AppTx
type AppTxReader struct {
	ChainID string
}

func (t AppTxReader) ReadTxJSON(data []byte, pk crypto.PubKey) (interface{}, error) {
	return nil, errors.New("Not implemented...")
}

func (t AppTxReader) ReadTxFlags(data *AppFlags, app string, appData []byte, pk crypto.PubKey) (interface{}, error) {
	//parse the fee and amounts into coin types
	feeCoin, err := btypes.ParseCoin(data.Fee)
	if err != nil {
		return nil, err
	}
	amountCoins, err := btypes.ParseCoins(data.Amount)
	if err != nil {
		return nil, err
	}

	// get addr if available
	var addr []byte
	if !pk.Empty() {
		addr = pk.Address()
	}

	// craft the tx
	input := btypes.TxInput{
		Address:  addr,
		Coins:    amountCoins,
		Sequence: data.Sequence,
	}
	if data.Sequence == 1 {
		input.PubKey = pk
	}
	tx := btypes.AppTx{
		Gas:   data.Gas,
		Fee:   feeCoin,
		Input: input,
		Name:  app,
		Data:  appData,
	}

	// wrap it in the proper signer thing...
	send := AppTx{
		chainID: t.ChainID,
		Tx:      &tx,
	}
	return &send, nil
}

/** TODO copied from basecoin cli - put in common somewhere? **/

// Returns true for non-empty hex-string prefixed with "0x"
func isHex(s string) bool {
	if len(s) > 2 && s[:2] == "0x" {
		_, err := hex.DecodeString(s[2:])
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func StripHex(s string) string {
	if isHex(s) {
		return s[2:]
	}
	return s
}
