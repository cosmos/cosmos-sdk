package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/types"

	cmn "github.com/tendermint/go-common"
	client "github.com/tendermint/go-rpc/client"
	"github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var TxFlags = []cli.Flag{
	NodeFlag,
	ChainIDFlag,

	FromFlag,

	AmountFlag,
	GasFlag,
	FeeFlag,
	SeqFlag,
}

var (
	TxCmd = cli.Command{
		Name:      "tx",
		Usage:     "Create, sign, and broadcast a transaction",
		ArgsUsage: "",
		Subcommands: []cli.Command{
			SendTxCmd,
			AppTxCmd,
			IbcTxCmd,
		},
	}

	SendTxCmd = cli.Command{
		Name:      "send",
		Usage:     "a SendTx transaction, for sending tokens around",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdSendTx(c)
		},
		Flags: append(TxFlags, ToFlag),
	}

	AppTxCmd = cli.Command{
		Name:      "app",
		Usage:     "an AppTx transaction, for sending raw data to plugins",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdAppTx(c)
		},
		Flags: append(TxFlags, NameFlag, DataFlag),
		// Subcommands are dynamically registered with plugins as needed
		Subcommands: []cli.Command{},
	}
)

// Register a subcommand of TxCmd to craft transactions for plugins
func RegisterTxSubcommand(cmd cli.Command) {
	TxCmd.Subcommands = append(TxCmd.Subcommands, cmd)
}

func cmdSendTx(c *cli.Context) error {
	toHex := c.String("to")
	fromFile := c.String("from")
	amount := c.String("amount")
	gas := int64(c.Int("gas"))
	fee := c.String("fee")
	chainID := c.String("chain_id")

	// convert destination address to bytes
	to, err := hex.DecodeString(StripHex(toHex))
	if err != nil {
		return errors.New("To address is invalid hex: " + err.Error())
	}

	// load the priv key
	privKey := LoadKey(fromFile)

	// get the sequence number for the tx
	sequence, err := getSeq(c, privKey.Address)
	if err != nil {
		return err
	}

	//parse the fee and amounts into coin types
	feeCoin, err := ParseCoin(fee)
	if err != nil {
		return err
	}
	amountCoins, err := ParseCoins(amount)
	if err != nil {
		return err
	}

	// craft the tx
	input := types.NewTxInput(privKey.PubKey, amountCoins, sequence)
	output := newOutput(to, amountCoins)
	tx := &types.SendTx{
		Gas:     gas,
		Fee:     feeCoin,
		Inputs:  []types.TxInput{input},
		Outputs: []types.TxOutput{output},
	}

	// sign that puppy
	signBytes := tx.SignBytes(chainID)
	tx.Inputs[0].Signature = privKey.Sign(signBytes)

	fmt.Println("Signed SendTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	// broadcast the transaction to tendermint
	if _, _, err := broadcastTx(c, tx); err != nil {
		return err
	}
	return nil
}

func cmdAppTx(c *cli.Context) error {
	// convert data to bytes
	dataString := c.String("data")
	data := []byte(dataString)
	if isHex(dataString) {
		data, _ = hex.DecodeString(dataString)
	}
	name := c.String("name")
	return AppTx(c, name, data)
}

func AppTx(c *cli.Context, name string, data []byte) error {
	fromFile := c.String("from")
	amount := c.String("amount")
	fee := c.String("fee")
	gas := int64(c.Int("gas"))
	chainID := c.String("chain_id")

	privKey := tmtypes.LoadPrivValidator(fromFile)

	sequence, err := getSeq(c, privKey.Address)
	if err != nil {
		return err
	}

	//parse the fee and amounts into coin types
	feeCoin, err := ParseCoin(fee)
	if err != nil {
		return err
	}
	amountCoins, err := ParseCoins(amount)
	if err != nil {
		return err
	}

	input := types.NewTxInput(privKey.PubKey, amountCoins, sequence)
	tx := &types.AppTx{
		Gas:   gas,
		Fee:   feeCoin,
		Name:  name,
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = privKey.Sign(tx.SignBytes(chainID))

	fmt.Println("Signed AppTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	data, log, err := broadcastTx(c, tx)
	if err != nil {
		return err
	}
	fmt.Printf("Response: %X ; %s\n", data, log)

	return nil
}

// broadcast the transaction to tendermint
func broadcastTx(c *cli.Context, tx types.Tx) ([]byte, string, error) {
	tmResult := new(ctypes.TMResult)
	tmAddr := c.String("node")
	clientURI := client.NewClientURI(tmAddr)

	// Don't you hate having to do this?
	// How many times have I lost an hour over this trick?!
	txBytes := []byte(wire.BinaryBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))
	_, err := clientURI.Call("broadcast_tx_commit", map[string]interface{}{"tx": txBytes}, tmResult)
	if err != nil {
		return nil, "", errors.New(cmn.Fmt("Error on broadcast tx: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultBroadcastTxCommit)
	// if it fails check, we don't even get a delivertx back!
	if !res.CheckTx.Code.IsOK() {
		r := res.CheckTx
		return nil, "", errors.New(cmn.Fmt("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log))
	}
	if !res.DeliverTx.Code.IsOK() {
		r := res.DeliverTx
		return nil, "", errors.New(cmn.Fmt("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log))
	}
	return res.DeliverTx.Data, res.DeliverTx.Log, nil
}

// if the sequence flag is set, return it;
// else, fetch the account by querying the app and return the sequence number
func getSeq(c *cli.Context, address []byte) (int, error) {
	if c.IsSet("sequence") {
		return c.Int("sequence"), nil
	}
	tmAddr := c.String("node")
	acc, err := getAcc(tmAddr, address)
	if err != nil {
		return 0, err
	}
	return acc.Sequence + 1, nil
}

func newOutput(to []byte, amount types.Coins) types.TxOutput {
	return types.TxOutput{
		Address: to,
		Coins:   amount,
	}

}
