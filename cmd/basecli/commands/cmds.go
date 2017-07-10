package commands

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"
	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
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
	FlagExpires  = "expires"
	FlagSequence = "sequence"
)

func init() {
	flags := SendTxCmd.Flags()
	flags.String(FlagTo, "", "Destination address for the bits")
	flags.String(FlagAmount, "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	flags.String(FlagFee, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	flags.Uint64(FlagGas, 0, "Amount of gas for this transaction")
	flags.Int64(FlagExpires, 0, "Block height at which this tx expires")
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
	tx, err = WrapChainTx(tx)
	if err != nil {
		return err
	}

	// Note: this is single sig (no multi sig yet)
	stx := auth.NewSig(tx)

	// Sign if needed and post.  This it the work-horse
	bres, err := txcmd.SignAndPostTx(stx)
	if err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

// WrapChainTx will wrap the tx with a ChainTx from the standard flags
func WrapChainTx(tx basecoin.Tx) (res basecoin.Tx, err error) {
	expires := viper.GetInt64(FlagExpires)
	if expires < 0 {
		return res, errors.New("expires must be >= 0")
	}
	res = base.NewChainTx(commands.GetChainID(), uint64(expires), tx)
	return res, nil
}

func readSendTxFlags() (tx basecoin.Tx, err error) {
	// parse to address
	chain, to, err := parseChainAddress(viper.GetString(FlagTo))
	if err != nil {
		return tx, err
	}
	toAddr := auth.SigPerm(to)
	toAddr.ChainID = chain

	// //parse the fee and amounts into coin types
	// tx.Fee, err = btypes.ParseCoin(viper.GetString(FlagFee))
	// if err != nil {
	// 	return err
	// }
	// // set the gas
	// tx.Gas = viper.GetInt64(FlagGas)

	amountCoins, err := coin.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return tx, err
	}

	// this could be much cooler with multisig...
	var fromAddr basecoin.Actor
	signer := txcmd.GetSigner()
	if !signer.Empty() {
		fromAddr = auth.SigPerm(signer.Address())
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

/** TODO copied from basecoin cli - put in common somewhere? **/

// ParseHexFlag parses a flag string to byte array
func ParseHexFlag(flag string) ([]byte, error) {
	return hex.DecodeString(cmn.StripHex(viper.GetString(flag)))
}
