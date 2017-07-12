package coin

import (
	"fmt"

	"github.com/tendermint/basecoin"
)

func init() {
	basecoin.TxMapper.RegisterImplementation(SendTx{}, TypeSend, ByteSend)
}

// we reserve the 0x20-0x3f range for standard modules
const (
	ByteSend = 0x20
	TypeSend = NameCoin + "/send"
)

//-----------------------------------------------------------------------------

// TxInput - expected coin movement outputs, used with SendTx
type TxInput struct {
	Address basecoin.Actor `json:"address"`
	Coins   Coins          `json:"coins"`
}

// ValidateBasic - validate transaction input
func (txIn TxInput) ValidateBasic() error {
	if txIn.Address.App == "" {
		return ErrInvalidAddress()
	}
	// TODO: knowledge of app-specific codings?
	if len(txIn.Address.Address) == 0 {
		return ErrInvalidAddress()
	}
	if !txIn.Coins.IsValid() {
		return ErrInvalidCoins()
	}
	if !txIn.Coins.IsPositive() {
		return ErrInvalidCoins()
	}
	return nil
}

func (txIn TxInput) String() string {
	return fmt.Sprintf("TxInput{%v,%v}", txIn.Address, txIn.Coins)
}

// NewTxInput - create a transaction input, used with SendTx
func NewTxInput(addr basecoin.Actor, coins Coins) TxInput {
	input := TxInput{
		Address: addr,
		Coins:   coins,
	}
	return input
}

//-----------------------------------------------------------------------------

// TxOutput - expected coin movement output, used with SendTx
type TxOutput struct {
	Address basecoin.Actor `json:"address"`
	Coins   Coins          `json:"coins"`
}

// ValidateBasic - validate transaction output
func (txOut TxOutput) ValidateBasic() error {
	if txOut.Address.App == "" {
		return ErrInvalidAddress()
	}
	// TODO: knowledge of app-specific codings?
	if len(txOut.Address.Address) == 0 {
		return ErrInvalidAddress()
	}
	if !txOut.Coins.IsValid() {
		return ErrInvalidCoins()
	}
	if !txOut.Coins.IsPositive() {
		return ErrInvalidCoins()
	}
	return nil
}

func (txOut TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%X,%v}", txOut.Address, txOut.Coins)
}

// NewTxOutput - create a transaction output, used with SendTx
func NewTxOutput(addr basecoin.Actor, coins Coins) TxOutput {
	output := TxOutput{
		Address: addr,
		Coins:   coins,
	}
	return output
}

//-----------------------------------------------------------------------------

// SendTx - high level transaction of the coin module
// Satisfies: TxInner
type SendTx struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

var _ basecoin.Tx = NewSendTx(nil, nil)

// NewSendTx - construct arbitrary multi-in, multi-out sendtx
func NewSendTx(in []TxInput, out []TxOutput) basecoin.Tx {
	return SendTx{Inputs: in, Outputs: out}.Wrap()
}

// NewSendOneTx is a helper for the standard (?) case where there is exactly
// one sender and one recipient
func NewSendOneTx(sender, recipient basecoin.Actor, amount Coins) basecoin.Tx {
	in := []TxInput{{Address: sender, Coins: amount}}
	out := []TxOutput{{Address: recipient, Coins: amount}}
	return SendTx{Inputs: in, Outputs: out}.Wrap()
}

// ValidateBasic - validate the send transaction
func (tx SendTx) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(tx.Inputs) == 0 {
		return ErrNoInputs()
	}
	if len(tx.Outputs) == 0 {
		return ErrNoOutputs()
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut Coins
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
		return ErrInvalidCoins()
	}
	return nil
}

func (tx SendTx) String() string {
	return fmt.Sprintf("SendTx{%v->%v}", tx.Inputs, tx.Outputs)
}

// Wrap - used to satisfy TxInner
func (tx SendTx) Wrap() basecoin.Tx {
	return basecoin.Tx{tx}
}
