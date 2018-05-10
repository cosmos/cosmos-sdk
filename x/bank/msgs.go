package bank

import (
	"encoding/json"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

var _ bam.Msg = MsgSend{}

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(in []Input, out []Output) MsgSend {
	return MsgSend{Inputs: in, Outputs: out}
}

// Implements Msg.
func (msg MsgSend) Type() string { return "bank" } // TODO: "bank/send"

// Implements Msg.
func (msg MsgSend) ValidateBasic() sdk.Error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs(DefaultCodespace).Trace("")
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs(DefaultCodespace).Trace("")
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut bam.Coins
	for _, in := range msg.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err.Trace("")
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.Trace("")
		}
		totalOut = totalOut.Plus(out.Coins)
	}
	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return bam.ErrInvalidCoins(totalIn.String()).Trace("inputs and outputs don't match")
	}
	return nil
}

// Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgSend) GetSigners() []bam.Address {
	addrs := make([]bam.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

//----------------------------------------
// MsgIssue

// MsgIssue - high level transaction of the coin module
type MsgIssue struct {
	Banker  bam.Address `json:"banker"`
	Outputs []Output    `json:"outputs"`
}

// NewMsgIssue - construct arbitrary multi-in, multi-out send msg.
func NewMsgIssue(banker bam.Address, out []Output) MsgIssue {
	return MsgIssue{Banker: banker, Outputs: out}
}

// Implements Msg.
func (msg MsgIssue) Type() string { return "bank" } // TODO: "bank/issue"

// Implements Msg.
func (msg MsgIssue) ValidateBasic() sdk.Error {
	// XXX
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs(DefaultCodespace).Trace("")
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.Trace("")
		}
	}
	return nil
}

// Implements Msg.
func (msg MsgIssue) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgIssue) GetSigners() []bam.Address {
	return []bam.Address{msg.Banker}
}

//----------------------------------------
// Input

// Transaction Output
type Input struct {
	Address bam.Address `json:"address"`
	Coins   bam.Coins   `json:"coins"`
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() sdk.Error {
	if len(in.Address) == 0 {
		return bam.ErrInvalidAddress(in.Address.String())
	}
	if !in.Coins.IsValid() {
		return bam.ErrInvalidCoins(in.Coins.String())
	}
	if !in.Coins.IsPositive() {
		return bam.ErrInvalidCoins(in.Coins.String())
	}
	return nil
}

// NewInput - create a transaction input, used with MsgSend
func NewInput(addr bam.Address, coins bam.Coins) Input {
	input := Input{
		Address: addr,
		Coins:   coins,
	}
	return input
}

//----------------------------------------
// Output

// Transaction Output
type Output struct {
	Address bam.Address `json:"address"`
	Coins   bam.Coins   `json:"coins"`
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() sdk.Error {
	if len(out.Address) == 0 {
		return bam.ErrInvalidAddress(out.Address.String())
	}
	if !out.Coins.IsValid() {
		return bam.ErrInvalidCoins(out.Coins.String())
	}
	if !out.Coins.IsPositive() {
		return bam.ErrInvalidCoins(out.Coins.String())
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgSend
func NewOutput(addr bam.Address, coins bam.Coins) Output {
	output := Output{
		Address: addr,
		Coins:   coins,
	}
	return output
}
