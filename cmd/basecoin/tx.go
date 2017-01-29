package main

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

func cmdSendTx(c *cli.Context) error {
	toHex := c.String("to")
	fromFile := c.String("from")
	amount := int64(c.Int("amount"))
	coin := c.String("coin")
	gas, fee := c.Int("gas"), int64(c.Int("fee"))
	chainID := c.String("chain_id")

	// convert destination address to bytes
	to, err := hex.DecodeString(toHex)
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
	if err := broadcastTx(c, tx); err != nil {
		return err
	}

	return nil
}

func cmdAppTx(c *cli.Context) error {
	name := c.String("name")
	fromFile := c.String("from")
	amount := int64(c.Int("amount"))
	coin := c.String("coin")
	gas, fee := c.Int("gas"), int64(c.Int("fee"))
	chainID := c.String("chain_id")
	dataString := c.String("data")

	// convert data to bytes
	data := []byte(dataString)
	if cmn.IsHex(dataString) {
		data, _ = hex.DecodeString(dataString)
	}

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

	if err := broadcastTx(c, tx); err != nil {
		return err
	}

	return nil
}

// broadcast the transaction to tendermint
func broadcastTx(c *cli.Context, tx types.Tx) error {
	tmResult := new(ctypes.TMResult)
	tmAddr := c.String("tendermint")
	clientURI := client.NewClientURI(tmAddr)

	/*txBytes := []byte(wire.JSONBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))*/
	txBytes := wire.BinaryBytes(tx)
	_, err := clientURI.Call("broadcast_tx_sync", map[string]interface{}{"tx": txBytes}, tmResult)
	if err != nil {
		return errors.New(cmn.Fmt("Error on broadcast tx: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultBroadcastTx)
	if !res.Code.IsOK() {
		return errors.New(cmn.Fmt("BroadcastTxSync got non-zero exit code: %v. %X; %s", res.Code, res.Data, res.Log))
	}
	return nil
}

// if the sequence flag is set, return it;
// else, fetch the account by querying the app and return the sequence number
func getSeq(c *cli.Context, address []byte) (int, error) {
	if c.IsSet("sequence") {
		return c.Int("sequence"), nil
	}
	tmAddr := c.String("tendermint")
	clientURI := client.NewClientURI(tmAddr)
	tmResult := new(ctypes.TMResult)

	params := map[string]interface{}{
		"path":  "/key",
		"data":  append([]byte("base/a/"), address...),
		"prove": false,
	}
	_, err := clientURI.Call("abci_query", params, tmResult)
	if err != nil {
		return 0, errors.New(cmn.Fmt("Error calling /abci_query: %v", err))
	}
	res := (*tmResult).(*ctypes.ResultABCIQuery)
	if !res.Response.Code.IsOK() {
		return 0, errors.New(cmn.Fmt("Query got non-zero exit code: %v. %s", res.Response.Code, res.Response.Log))
	}
	accountBytes := res.Response.Value

	if len(accountBytes) == 0 {
		return 0, errors.New(cmn.Fmt("Account bytes are empty from query for address %X", address))
	}
	var acc *types.Account
	err = wire.ReadBinaryBytes(accountBytes, &acc)
	if err != nil {
		return 0, errors.New(cmn.Fmt("Error reading account %X error: %v",
			accountBytes, err.Error()))
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
