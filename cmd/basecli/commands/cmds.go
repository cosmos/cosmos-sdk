package commands

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	cmn "github.com/tendermint/tmlibs/common"

	btypes "github.com/tendermint/basecoin/types"
)

//-------------------------
// SendTx

// SendTxCmd is CLI command to send tokens between basecoin accounts
var SendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send tokens from one account to another",
	RunE:  doSendTx,
}

//nolint
const (
	FlagTo       = "to"
	FlagAmount   = "amount"
	FlagFee      = "fee"
	FlagGas      = "gas"
	FlagSequence = "sequence"
)

func init() {
	flags := SendTxCmd.Flags()
	flags.String(FlagTo, "", "Destination address for the bits")
	flags.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	flags.String(FlagFee, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	flags.Int64(FlagGas, 0, "Amount of gas for this transaction")
	flags.Int(FlagSequence, -1, "Sequence number for this transaction")
}

// runDemo is an example of how to make a tx
func doSendTx(cmd *cobra.Command, args []string) error {

	// load data from json or flags
	tx := new(btypes.SendTx)
	found, err := txcmd.LoadJSON(tx)
	if err != nil {
		return err
	}
	if !found {
		err = readSendTxFlags(tx)
	}
	if err != nil {
		return err
	}

	// Wrap and add signer
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

	// Output result
	return txcmd.OutputTx(bres)
}

func readSendTxFlags(tx *btypes.SendTx) error {
	// parse to address
	to, err := parseChainAddress(viper.GetString(FlagTo))
	if err != nil {
		return err
	}

	//parse the fee and amounts into coin types
	tx.Fee, err = btypes.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return err
	}
	amountCoins, err := btypes.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	// set the gas
	tx.Gas = viper.GetInt64(FlagGas)

	// craft the inputs and outputs
	tx.Inputs = []btypes.TxInput{{
		Coins:    amountCoins,
		Sequence: viper.GetInt(FlagSequence),
	}}
	tx.Outputs = []btypes.TxOutput{{
		Address: to,
		Coins:   amountCoins,
	}}

	return nil
}

func parseChainAddress(toFlag string) ([]byte, error) {
	var toHex string
	var chainPrefix string
	spl := strings.Split(toFlag, "/")
	switch len(spl) {
	case 1:
		toHex = spl[0]
	case 2:
		chainPrefix = spl[0]
		toHex = spl[1]
	default:
		return nil, errors.Errorf("To address has too many slashes")
	}

	// convert destination address to bytes
	to, err := hex.DecodeString(cmn.StripHex(toHex))
	if err != nil {
		return nil, errors.Errorf("To address is invalid hex: %v\n", err)
	}

	if chainPrefix != "" {
		to = []byte(chainPrefix + "/" + string(to))
	}
	return to, nil
}

//-------------------------
// AppTx

// BroadcastAppTx wraps, signs, and executes an app tx basecoin transaction
func BroadcastAppTx(tx *btypes.AppTx) (*ctypes.ResultBroadcastTxCommit, error) {

	// Generate app transaction to be broadcast
	appTx := WrapAppTx(tx)
	appTx.AddSigner(txcmd.GetSigner())

	// Sign if needed and post to the node.  This it the work-horse
	return txcmd.SignAndPostTx(appTx)
}

// AddAppTxFlags adds flags required by apptx
func AddAppTxFlags(fs *flag.FlagSet) {
	fs.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.String(FlagFee, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	fs.Int64(FlagGas, 0, "Amount of gas for this transaction")
	fs.Int(FlagSequence, -1, "Sequence number for this transaction")
}

// ReadAppTxFlags reads in the standard flags
// your command should parse info to set tx.Name and tx.Data
func ReadAppTxFlags() (gas int64, fee btypes.Coin, txInput btypes.TxInput, err error) {

	// Set the gas
	gas = viper.GetInt64(FlagGas)

	// Parse the fee and amounts into coin types
	fee, err = btypes.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return
	}

	// craft the inputs
	var amount btypes.Coins
	amount, err = btypes.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return
	}
	txInput = btypes.TxInput{
		Coins:    amount,
		Sequence: viper.GetInt(FlagSequence),
	}

	return
}

// WrapAppTx wraps the transaction with chain id
func WrapAppTx(tx *btypes.AppTx) *AppTx {
	return &AppTx{
		chainID: commands.GetChainID(),
		Tx:      tx,
	}
}

/** TODO copied from basecoin cli - put in common somewhere? **/

// ParseHexFlag parses a flag string to byte array
func ParseHexFlag(flag string) ([]byte, error) {
	return hex.DecodeString(cmn.StripHex(viper.GetString(flag)))
}
