package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/commands"
	txcmd "github.com/tendermint/light-client/commands/txs"
	cmn "github.com/tendermint/tmlibs/common"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/nonce"
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
	flags.Uint64(FlagExpires, 0, "Block height at which this tx expires")
	flags.Int(FlagSequence, -1, "Sequence number for this transaction")
}

// doSendTx is an example of how to make a tx
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
	tx, err = WrapFeeTx(tx)
	if err != nil {
		return err
	}
	tx, err = WrapNonceTx(tx)
	if err != nil {
		return err
	}
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
	if err = ValidateResult(bres); err != nil {
		return err
	}

	// Output result
	return txcmd.OutputTx(bres)
}

// ValidateResult returns an appropriate error if the server rejected the
// tx in CheckTx or DeliverTx
func ValidateResult(res *ctypes.ResultBroadcastTxCommit) error {
	if res.CheckTx.IsErr() {
		return fmt.Errorf("CheckTx: (%d): %s", res.CheckTx.Code, res.CheckTx.Log)
	}
	if res.DeliverTx.IsErr() {
		return fmt.Errorf("DeliverTx: (%d): %s", res.DeliverTx.Code, res.DeliverTx.Log)
	}
	return nil
}

// WrapNonceTx grabs the sequence number from the flag and wraps
// the tx with this nonce.  Grabs the permission from the signer,
// as we still only support single sig on the cli
func WrapNonceTx(tx basecoin.Tx) (res basecoin.Tx, err error) {
	//add the nonce tx layer to the tx
	seq := viper.GetInt(FlagSequence)
	if seq < 0 {
		return res, fmt.Errorf("sequence must be greater than 0")
	}
	signers := []basecoin.Actor{GetSignerAct()}
	tx = nonce.NewTx(uint32(seq), signers, tx)
	return tx, nil
}

// WrapFeeTx checks for FlagFee and if present wraps the tx with a
// FeeTx of the given amount, paid by the signer
func WrapFeeTx(tx basecoin.Tx) (res basecoin.Tx, err error) {
	//parse the fee and amounts into coin types
	toll, err := coin.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return res, err
	}
	// if no fee, do nothing, otherwise wrap it
	if toll.IsZero() {
		return tx, nil
	}
	return fee.NewFee(tx, toll, GetSignerAct()), nil
}

// WrapChainTx will wrap the tx with a ChainTx from the standard flags
func WrapChainTx(tx basecoin.Tx) (res basecoin.Tx, err error) {
	expires := viper.GetInt64(FlagExpires)
	chain := commands.GetChainID()
	if chain == "" {
		return res, errors.New("No chain-id provided")
	}
	res = base.NewChainTx(chain, uint64(expires), tx)
	return res, nil
}

// GetSignerAct returns the address of the signer of the tx
// (as we still only support single sig)
func GetSignerAct() (res basecoin.Actor) {
	// this could be much cooler with multisig...
	signer := txcmd.GetSigner()
	if !signer.Empty() {
		res = auth.SigPerm(signer.Address())
	}
	return res
}

func readSendTxFlags() (tx basecoin.Tx, err error) {
	// parse to address
	chain, to, err := parseChainAddress(viper.GetString(FlagTo))
	if err != nil {
		return tx, err
	}
	toAddr := auth.SigPerm(to)
	toAddr.ChainID = chain

	amountCoins, err := coin.ParseCoins(viper.GetString(FlagAmount))
	if err != nil {
		return tx, err
	}

	// craft the inputs and outputs
	ins := []coin.TxInput{{
		Address: GetSignerAct(),
		Coins:   amountCoins,
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
