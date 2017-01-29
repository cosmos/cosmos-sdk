package main

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

func cmdSendTx(c *cli.Context) error {
	toHex := c.String("to")
	fromFile := c.String("from")
	amount := int64(c.Int("amount"))
	coin := c.String("coin")
	gas, fee := c.Int("gas"), int64(c.Int("fee"))
	chainID := c.String("chain_id")

	to, err := hex.DecodeString(toHex)
	if err != nil {
		return errors.New("To address is invalid hex: " + err.Error())
	}

	privVal := tmtypes.LoadPrivValidator(fromFile)

	sequence := getSeq(c)

	input := types.NewTxInput(privVal.PubKey, types.Coins{types.Coin{coin, amount}}, sequence)
	output := newOutput(to, coin, amount)

	tx := types.SendTx{
		Gas:     int64(gas),
		Fee:     types.Coin{coin, fee},
		Inputs:  []types.TxInput{input},
		Outputs: []types.TxOutput{output},
	}

	tx.Inputs[0].Signature = privVal.Sign(tx.SignBytes(chainID))
	fmt.Println(string(wire.JSONBytes(tx)))

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

	data := []byte(dataString)
	if cmn.IsHex(dataString) {
		data, _ = hex.DecodeString(dataString)
	}

	privVal := tmtypes.LoadPrivValidator(fromFile)

	sequence := getSeq(c)

	input := types.NewTxInput(privVal.PubKey, types.Coins{types.Coin{coin, amount}}, sequence)

	tx := types.AppTx{
		Gas:   int64(gas),
		Fee:   types.Coin{coin, fee},
		Name:  name,
		Input: input,
		Data:  data,
	}

	tx.Input.Signature = privVal.Sign(tx.SignBytes(chainID))
	fmt.Println(string(wire.JSONBytes(tx)))
	return nil
}

func getSeq(c *cli.Context) int {
	if c.IsSet("sequence") {
		return c.Int("sequence")
	}
	// TODO: get from query
	return 0
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
