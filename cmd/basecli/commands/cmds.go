package commands

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

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
		chainID: commands.GetChainID(),
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
	to, err := ParseHexFlag(ToFlag)
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

func AddAppTxFlags(fs *flag.FlagSet) {
	fs.String(AmountFlag, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.String(FeeFlag, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	fs.Int64(GasFlag, 0, "Amount of gas for this transaction")
	fs.Int(SequenceFlag, -1, "Sequence number for this transaction")
}

// ReadAppTxFlags reads in the standard flags
// your command should parse info to set tx.Name and tx.Data
func ReadAppTxFlags(tx *btypes.AppTx) error {
	//parse the fee and amounts into coin types
	var err error
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
	tx.Input = btypes.TxInput{
		Coins:    amountCoins,
		Sequence: viper.GetInt(SequenceFlag),
	}

	return nil
}

func WrapAppTx(tx *btypes.AppTx) *AppTx {
	return &AppTx{
		chainID: commands.GetChainID(),
		Tx:      tx,
	}
}

/** TODO copied from basecoin cli - put in common somewhere? **/

func ParseHexFlag(flag string) ([]byte, error) {
	return hex.DecodeString(StripHex(viper.GetString(flag)))
}

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
