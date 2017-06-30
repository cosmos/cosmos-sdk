package coin

import (
	"fmt"

	"github.com/tendermint/basecoin"

	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
)

//-----------------------------------------------------------------------------

type TxInput struct {
	Address  basecoin.Actor `json:"address"`
	Coins    types.Coins    `json:"coins"`
	Sequence int            `json:"sequence"` // Nonce: Must be 1 greater than the last committed TxInput
}

func (txIn TxInput) ValidateBasic() error {
	if txIn.Address.App == "" {
		return errors.InvalidAddress()
	}
	// TODO: knowledge of app-specific codings?
	if len(txIn.Address.Address) == 0 {
		return errors.InvalidAddress()
	}
	if !txIn.Coins.IsValid() {
		return errors.InvalidCoins()
	}
	if !txIn.Coins.IsPositive() {
		return errors.InvalidCoins()
	}
	if txIn.Sequence <= 0 {
		return errors.InvalidSequence()
	}
	return nil
}

func (txIn TxInput) String() string {
	return fmt.Sprintf("TxInput{%v,%v,%v}", txIn.Address, txIn.Coins, txIn.Sequence)
}

func NewTxInput(addr basecoin.Actor, coins types.Coins, sequence int) TxInput {
	input := TxInput{
		Address:  addr,
		Coins:    coins,
		Sequence: sequence,
	}
	return input
}

//-----------------------------------------------------------------------------

type TxOutput struct {
	Address basecoin.Actor `json:"address"`
	Coins   types.Coins    `json:"coins"`
}

func (txOut TxOutput) ValidateBasic() error {
	if txOut.Address.App == "" {
		return errors.InvalidAddress()
	}
	// TODO: knowledge of app-specific codings?
	if len(txOut.Address.Address) == 0 {
		return errors.InvalidAddress()
	}
	if !txOut.Coins.IsValid() {
		return errors.InvalidCoins()
	}
	if !txOut.Coins.IsPositive() {
		return errors.InvalidCoins()
	}
	return nil
}

func (txOut TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%X,%v}", txOut.Address, txOut.Coins)
}

func NewTxOutput(addr basecoin.Actor, coins types.Coins) TxOutput {
	output := TxOutput{
		Address: addr,
		Coins:   coins,
	}
	return output
}

//-----------------------------------------------------------------------------

type SendTx struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

var _ basecoin.Tx = SendTx{}.Wrap()

func (tx SendTx) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(tx.Inputs) == 0 {
		return errors.NoInputs()
	}
	if len(tx.Outputs) == 0 {
		return errors.NoOutputs()
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut types.Coins
	for _, in := range tx.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range tx.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err
		}
		totalOut = totalOut.Plus(out.Coins)
	}
	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return errors.InvalidCoins()
	}
	return nil
}

func (tx SendTx) String() string {
	return fmt.Sprintf("SendTx{%v->%v}", tx.Inputs, tx.Outputs)
}

func (tx SendTx) Wrap() basecoin.Tx {
	return basecoin.Tx{tx}
}
