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

var (
	SendTxCmd = cli.Command{
		Name:      "sendtx",
		Usage:     "Broadcast a basecoin SendTx",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdSendTx(c)
		},
		Flags: []cli.Flag{
			NodeFlag,
			ChainIDFlag,

			FromFlag,

			AmountFlag,
			CoinFlag,
			GasFlag,
			FeeFlag,
			SeqFlag,

			ToFlag,
		},
	}

	AppTxCmd = cli.Command{
		Name:      "apptx",
		Usage:     "Broadcast a basecoin AppTx",
		ArgsUsage: "",
		Action: func(c *cli.Context) error {
			return cmdAppTx(c)
		},
		Flags: []cli.Flag{
			NodeFlag,
			ChainIDFlag,

			FromFlag,

			AmountFlag,
			CoinFlag,
			GasFlag,
			FeeFlag,
			SeqFlag,

			NameFlag,
			DataFlag,
		},
		// Subcommands are dynamically registered with plugins as needed
		Subcommands: []cli.Command{},
	}
)

// RegisterPlugin is used to add another subcommand and create a custom
// apptx encoding.  Look at counter.go for an example
func RegisterPlugin(cmd cli.Command) {
	AppTxCmd.Subcommands = append(AppTxCmd.Subcommands, cmd)
}

func cmdSendTx(c *cli.Context) error {
	toHex := c.String("to")
	fromFile := c.String("from")
	amount := int64(c.Int("amount"))
	coin := c.String("coin")
	gas, fee := c.Int("gas"), int64(c.Int("fee"))
	chainID := c.String("chain_id")

	// convert destination address to bytes
	to, err := hex.DecodeString(stripHex(toHex))
	if err != nil {
		return errors.New("To address is invalid hex: " + err.Error())
	}

	// load the priv validator
	// XXX: this is overkill for now, we need a keys solution
	privVal := tmtypes.LoadPrivValidator(fromFile)

	// get the sequence number for the tx
	sequence, err := getSeq(c, privVal.Address)
	if err != nil {
		return err
	}

	// craft the tx
	input := types.NewTxInput(privVal.PubKey, types.Coins{types.Coin{coin, amount}}, sequence)
	output := newOutput(to, coin, amount)
	tx := &types.SendTx{
		Gas:     int64(gas),
		Fee:     types.Coin{coin, fee},
		Inputs:  []types.TxInput{input},
		Outputs: []types.TxOutput{output},
	}

	// sign that puppy
	signBytes := tx.SignBytes(chainID)
	tx.Inputs[0].Signature = privVal.Sign(signBytes)

	fmt.Println("Signed SendTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	// broadcast the transaction to tendermint
	if _, err := broadcastTx(c, tx); err != nil {
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
	amount := int64(c.Int("amount"))
	coin := c.String("coin")
	gas, fee := c.Int("gas"), int64(c.Int("fee"))
	chainID := c.String("chain_id")

	privVal := tmtypes.LoadPrivValidator(fromFile)

	sequence, err := getSeq(c, privVal.Address)
	if err != nil {
		return err
	}

	input := types.NewTxInput(privVal.PubKey, types.Coins{types.Coin{coin, amount}}, sequence)
	tx := &types.AppTx{
		Gas:   int64(gas),
		Fee:   types.Coin{coin, fee},
		Name:  name,
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = privVal.Sign(tx.SignBytes(chainID))

	fmt.Println("Signed AppTx:")
	fmt.Println(string(wire.JSONBytes(tx)))

	if _, err := broadcastTx(c, tx); err != nil {
		return err
	}

	return nil
}

// broadcast the transaction to tendermint
func broadcastTx(c *cli.Context, tx types.Tx) ([]byte, error) {
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
		return nil, errors.New(cmn.Fmt("Error on broadcast tx: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultBroadcastTxCommit)
	if !res.DeliverTx.Code.IsOK() {
		r := res.DeliverTx
		return nil, errors.New(cmn.Fmt("BroadcastTxCommit got non-zero exit code: %v. %X; %s", r.Code, r.Data, r.Log))
	}
	return res.DeliverTx.Data, nil
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

func newOutput(to []byte, coin string, amount int64) types.TxOutput {
	return types.TxOutput{
		Address: to,
		Coins: types.Coins{
			types.Coin{
				Denom:  coin,
				Amount: amount,
			},
		},
	}

}
