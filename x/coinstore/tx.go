package coinstore

import (
	"encoding/json"
	"fmt"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/coin"
)

//-----------------------------------------------------------------------------

// TxInput
type TxInput struct {
	Address  crypto.Address `json:"address"`
	Coins    Coins          `json:"coins"`
	Sequence int64          `json:"sequence"`

	signature crypto.Signature
}

// ValidateBasic - validate transaction input
func (txIn TxInput) ValidateBasic() error {
	if len(txIn.Address) == 0 {
		return ErrInvalidAddress(txIn.Address.String())
	}
	if txIn.Sequence < 0 {
		return ErrInvalidSequence(txIn.Sequence)
	}
	if !txIn.Coins.IsValid() {
		return ErrInvalidCoins(txIn.Coins.String())
	}
	if !txIn.Coins.IsPositive() {
		return ErrInvalidCoins(txIn.Coins.String())
	}
	return nil
}

func (txIn TxInput) String() string {
	return fmt.Sprintf("TxInput{%v,%v}", txIn.Address, txIn.Coins)
}

// NewTxInput - create a transaction input, used with SendTx
func NewTxInput(addr crypto.Address, coins Coins) TxInput {
	input := TxInput{
		Address: addr,
		Coins:   coins,
	}
	return input
}

// NewTxInputWithSequence - create a transaction input, used with SendTx
func NewTxInputWithSequence(addr crypto.Address, coins Coins, seq int64) TxInput {
	input := NewTxInput(addr, coins)
	input.Sequence = seq
	return input
}

//-----------------------------------------------------------------------------

// TxOutput - expected coin movement output, used with SendTx
type TxOutput struct {
	Address crypto.Address `json:"address"`
	Coins   Coins          `json:"coins"`
}

// ValidateBasic - validate transaction output
func (txOut TxOutput) ValidateBasic() error {
	if len(txOut.Address) == 0 {
		return ErrInvalidAddress(txOut.Address.String())
	}
	if !txOut.Coins.IsValid() {
		return ErrInvalidCoins(txOut.Coins.String())
	}
	if !txOut.Coins.IsPositive() {
		return ErrInvalidCoins(txOut.Coins.String())
	}
	return nil
}

func (txOut TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%X,%v}", txOut.Address, txOut.Coins)
}

// NewTxOutput - create a transaction output, used with SendTx
func NewTxOutput(addr crypto.Address, coins Coins) TxOutput {
	output := TxOutput{
		Address: addr,
		Coins:   coins,
	}
	return output
}

//-----------------------------------------------------------------------------

var _ types.Tx = (*SendTx)(nil)

// SendTx - high level transaction of the coin module
type SendTx struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
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
		return ErrInvalidCoins(totalIn.String()) // TODO
	}
	return nil
}

func (tx SendTx) String() string {
	return fmt.Sprintf("SendTx{%v->%v}", tx.Inputs, tx.Outputs)
}

// NewSendTx - construct arbitrary multi-in, multi-out sendtx
func NewSendTx(in []TxInput, out []TxOutput) types.Tx {
	return SendTx{Inputs: in, Outputs: out}
}

// NewSendOneTx is a helper for the standard (?) case where there is exactly
// one sender and one recipient
func NewSendOneTx(sender, recipient crypto.Address, amount coin.Coins) types.Tx {
	in := []TxInput{{Address: sender, Coins: amount}}
	out := []TxOutput{{Address: recipient, Coins: amount}}
	return SendTx{Inputs: in, Outputs: out}
}

//------------------------
// Implements types.Tx

func (tx SendTx) Get(key interface{}) (value interface{}) {
	switch k := key.(type) {
	case string:
		switch k {
		case "key":
		case "value":
		}
	}
	return nil
}

func (tx SendTx) SignBytes() []byte {
	b, err := json.Marshal(tx) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

func (tx SendTx) Signers() []crypto.Address {
	addrs := make([]crypto.Address, len(tx.Inputs))
	for i, in := range tx.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

func (tx SendTx) TxBytes() []byte {
	b, err := json.Marshal(struct {
		Tx        types.Tx           `json:"tx"`
		Signature []crypto.Signature `json:"signature"`
	}{
		Tx:        tx,
		Signature: tx.signatures(),
	})
	if err != nil {
		panic(err)
	}
	return b
}

func (tx SendTx) Signatures() []types.StdSignature {
	stdSigs := make([]types.StdSignature, len(tx.Inputs))
	for i, in := range tx.Inputs {
		stdSigs[i] = types.StdSignature{
			Signature: in.signature,
			Sequence:  in.Sequence,
		}
	}
	return stdSigs
}

func (tx SendTx) signatures() []crypto.Signature {
	sigs := make([]crypto.Signature, len(tx.Inputs))
	for i, in := range tx.Inputs {
		sigs[i] = in.signature
	}
	return sigs
}
