package main

import (
	"encoding/hex"
	"encoding/json"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	btypes "github.com/tendermint/basecoin/types"
	keycmd "github.com/tendermint/go-crypto/cmd"
	wire "github.com/tendermint/go-wire"
	lightclient "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/commands"
	"github.com/tendermint/light-client/proofs"
)

type AccountPresenter struct{}

func (_ AccountPresenter) MakeKey(str string) ([]byte, error) {
	res, err := hex.DecodeString(str)
	if err == nil {
		res = append([]byte("base/a/"), res...)
	}
	return res, err
}

func (_ AccountPresenter) ParseData(raw []byte) (interface{}, error) {
	var acc *btypes.Account
	err := wire.ReadBinaryBytes(raw, &acc)
	return acc, err
}

type BaseTxPresenter struct {
	proofs.RawPresenter // this handles MakeKey as hex bytes
}

func (_ BaseTxPresenter) ParseData(raw []byte) (interface{}, error) {
	var tx btypes.TxS
	err := wire.ReadBinaryBytes(raw, &tx)
	return tx, err
}

// SendTXReader allows us to create SendTx
type SendTxReader struct {
	ChainID string
}

func (t SendTxReader) ReadTxJSON(data []byte) (interface{}, error) {
	var tx btypes.SendTx
	err := json.Unmarshal(data, &tx)
	send := SendTx{
		chainID: t.ChainID,
		Tx:      &tx,
	}
	return &send, errors.Wrap(err, "parse sendtx")
}

type SendTxMaker struct{}

func (m SendTxMaker) MakeReader() (lightclient.TxReader, error) {
	chainID := viper.GetString(commands.ChainFlag)
	return SendTxReader{ChainID: chainID}, nil
}

type SendFlags struct {
	To       string
	Amount   string
	Fee      string
	Gas      int64
	Sequence int
}

func (m SendTxMaker) Flags() (*flag.FlagSet, interface{}) {
	fs := flag.NewFlagSet("foobar", flag.ContinueOnError)
	fs.String("to", "", "Destination address for the bits")
	fs.String("amount", "", "Coins to send in the format <amt><coin>,<amt><coin>...")
	fs.String("fee", "", "Coins for the transaction fee of the format <amt><coin>")
	fs.Int64("gas", 0, "Amount of gas for this transaction")
	fs.Int("sequence", -1, "Sequence number for this transaction")
	return fs, &SendFlags{}
}

func (t SendTxReader) ReadTxFlags(flags interface{}) (interface{}, error) {
	data := flags.(*SendFlags)

	// parse to and from addresses
	to, err := hex.DecodeString(StripHex(data.To))
	if err != nil {
		return nil, errors.Errorf("To address is invalid hex: %v\n", err)
	}

	// TODO: figure out a cleaner way to do this... until then
	// just close your eyes and continue...
	manager := keycmd.GetKeyManager()
	name := viper.GetString("name")
	info, err := manager.Get(name)

	//parse the fee and amounts into coin types
	feeCoin, err := btypes.ParseCoin(data.Fee)
	if err != nil {
		return nil, err
	}
	amountCoins, err := btypes.ParseCoins(data.Amount)
	if err != nil {
		return nil, err
	}

	// craft the tx
	input := btypes.TxInput{
		Address:  info.Address,
		Coins:    amountCoins,
		Sequence: data.Sequence,
	}
	if data.Sequence == 1 {
		input.PubKey = info.PubKey
	}
	output := btypes.TxOutput{
		Address: to,
		Coins:   amountCoins,
	}
	tx := btypes.SendTx{
		Gas:     data.Gas,
		Fee:     feeCoin,
		Inputs:  []btypes.TxInput{input},
		Outputs: []btypes.TxOutput{output},
	}

	// wrap it in the proper signer thing...
	send := SendTx{
		chainID: t.ChainID,
		Tx:      &tx,
	}
	return &send, nil
}

/** copied from basecoin cli - put in common somewhere? **/

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
