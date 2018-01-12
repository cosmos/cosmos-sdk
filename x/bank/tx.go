package bank

import (
	"encoding/json"
	"fmt"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/types"
)

// SendMsg - high level transaction of the coin module
type SendMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

// NewSendMsg - construct arbitrary multi-in, multi-out send msg.
func NewSendMsg(in []Input, out []Output) SendMsg {
	return SendMsg{Inputs: in, Outputs: out}
}

// Implements Msg.
func (msg SendMsg) Type() string { return "bank" } // TODO: "bank/send"

// Implements Msg.
func (msg SendMsg) ValidateBasic() error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs()
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs()
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut types.Coins
	for _, in := range msg.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range msg.Outputs {
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

func (msg SendMsg) String() string {
	return fmt.Sprintf("SendMsg{%v->%v}", msg.Inputs, msg.Outputs)
}

// Implements Msg.
func (msg SendMsg) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg SendMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg SendMsg) GetSigners() []crypto.Address {
	addrs := make([]crypto.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

//----------------------------------------
// Input

type Input struct {
	Address  crypto.Address `json:"address"`
	Coins    types.Coins    `json:"coins"`
	Sequence int64          `json:"sequence"`

	signature crypto.Signature
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() error {
	if len(in.Address) == 0 {
		return ErrInvalidAddress(in.Address.String())
	}
	if in.Sequence < 0 {
		return ErrInvalidSequence(in.Sequence)
	}
	if !in.Coins.IsValid() {
		return ErrInvalidCoins(in.Coins.String())
	}
	if !in.Coins.IsPositive() {
		return ErrInvalidCoins(in.Coins.String())
	}
	return nil
}

func (in Input) String() string {
	return fmt.Sprintf("Input{%v,%v}", in.Address, in.Coins)
}

// NewInput - create a transaction input, used with SendMsg
func NewInput(addr crypto.Address, coins types.Coins) Input {
	input := Input{
		Address: addr,
		Coins:   coins,
	}
	return input
}

// NewInputWithSequence - create a transaction input, used with SendMsg
func NewInputWithSequence(addr crypto.Address, coins types.Coins, seq int64) Input {
	input := NewInput(addr, coins)
	input.Sequence = seq
	return input
}

//----------------------------------------
// Output

type Output struct {
	Address crypto.Address `json:"address"`
	Coins   types.Coins    `json:"coins"`
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() error {
	if len(out.Address) == 0 {
		return ErrInvalidAddress(out.Address.String())
	}
	if !out.Coins.IsValid() {
		return ErrInvalidCoins(out.Coins.String())
	}
	if !out.Coins.IsPositive() {
		return ErrInvalidCoins(out.Coins.String())
	}
	return nil
}

func (out Output) String() string {
	return fmt.Sprintf("Output{%X,%v}", out.Address, out.Coins)
}

// NewOutput - create a transaction output, used with SendMsg
func NewOutput(addr crypto.Address, coins types.Coins) Output {
	output := Output{
		Address: addr,
		Coins:   coins,
	}
	return output
}
