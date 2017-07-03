package commands

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/txs"
	btypes "github.com/tendermint/basecoin/types"
)

//-------------------------
// SendTx

// SendTxCmd is CLI command to send tokens between basecoin accounts
var SendTxCmd = &cobra.Command{
	Use:   "send",
	Short: "send tokens from one account to another",
	RunE:  commands.RequireInit(doSendTx),
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
	var tx basecoin.Tx
	found, err := txcmd.LoadJSON(&tx)
	if err != nil {
		return err
	}
	if !found {
		tx, err = readSendTxFlags()
	}
	if err != nil {
		return err
	}

	// TODO: make this more flexible for middleware
	// add the chain info
	tx = txs.NewChain(commands.GetChainID(), tx)
	stx := txs.NewSig(tx)

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(stx)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

func readSendTxFlags() (tx basecoin.Tx, err error) {
	// parse to address
	chain, to, err := parseChainAddress(viper.GetString(FlagTo))
	if err != nil {
		return tx, err
	}
	toAddr := stack.SigPerm(to)
	toAddr.ChainID = chain

	// //parse the fee and amounts into coin types
	// tx.Fee, err = btypes.ParseCoin(viper.GetString(FlagFee))
	// if err != nil {
	// 	return err
	// }
	// // set the gas
	// tx.Gas = viper.GetInt64(FlagGas)

	amountCoins, err := btypes.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return tx, err
	}

	// this could be much cooler with multisig...
	var fromAddr basecoin.Actor
	signer := txcmd.GetSigner()
	if !signer.Empty() {
		fromAddr = stack.SigPerm(signer.Address())
	}

	// craft the inputs and outputs
	ins := []coin.TxInput{{
		Address:  fromAddr,
		Coins:    amountCoins,
		Sequence: viper.GetInt(FlagSequence),
	}}
	outs := []coin.TxOutput{{
		Address: toAddr,
		Coins:   amountCoins,
	}}

	return coin.NewSendTx(ins, outs), nil
}

func parseChainAddress(toFlag string) (string, []byte, error) {
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
		return "", nil, errors.Errorf("To address has too many slashes")
	}

	// convert destination address to bytes
	to, err := hex.DecodeString(cmn.StripHex(toHex))
	if err != nil {
		return "", nil, errors.Errorf("To address is invalid hex: %v\n", err)
	}

	return chainPrefix, to, nil
}

//-------------------------
// AppTx

// BroadcastAppTx wraps, signs, and executes an app tx basecoin transaction
func BroadcastAppTx(tx *btypes.AppTx) (*ctypes.ResultBroadcastTxCommit, error) {

	// Sign if needed and post to the node.  This it the work-horse
	return txcmd.SignAndPostTx(WrapAppTx(tx))
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

	// retrieve the amount
	var amount btypes.Coins
	amount, err = btypes.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return
	}

	// get the PubKey of the signer
	pk := txcmd.GetSigner()

	// get addr if available
	var addr []byte
	if !pk.Empty() {
		addr = pk.Address()
	}

	// set the output
	txInput = btypes.TxInput{
		Coins:    amount,
		Sequence: viper.GetInt(FlagSequence),
		Address:  addr,
	}
	// set the pubkey if needed
	if txInput.Sequence == 1 {
		txInput.PubKey = pk
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
