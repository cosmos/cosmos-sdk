package bank

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

var _ sdk.Msg = MsgSend{}

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(in []Input, out []Output) MsgSend {
	return MsgSend{Inputs: in, Outputs: out}
}

// Implements Msg.
// nolint
func (msg MsgSend) Route() string { return "bank" } // TODO: "bank/send"
func (msg MsgSend) Type() string  { return "send" }

// Implements Msg.
func (msg MsgSend) ValidateBasic() sdk.Error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs(DefaultCodespace).TraceSDK("")
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs(DefaultCodespace).TraceSDK("")
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut sdk.Coins
	for _, in := range msg.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalOut = totalOut.Plus(out.Coins)
	}
	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return sdk.ErrInvalidCoins(totalIn.String()).TraceSDK("inputs and outputs don't match")
	}
	return nil
}

// Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	var inputs, outputs []json.RawMessage
	for _, input := range msg.Inputs {
		inputs = append(inputs, input.GetSignBytes())
	}
	for _, output := range msg.Outputs {
		outputs = append(outputs, output.GetSignBytes())
	}
	b, err := msgCdc.MarshalJSON(struct {
		Inputs  []json.RawMessage `json:"inputs"`
		Outputs []json.RawMessage `json:"outputs"`
	}{
		Inputs:  inputs,
		Outputs: outputs,
	})
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

//----------------------------------------
// Input

// Transaction Input
type Input struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

// Return bytes to sign for Input
func (in Input) GetSignBytes() []byte {
	bin, err := msgCdc.MarshalJSON(in)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bin)
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() sdk.Error {
	if len(in.Address) == 0 {
		return sdk.ErrInvalidAddress(in.Address.String())
	}
	if !in.Coins.IsValid() {
		return sdk.ErrInvalidCoins(in.Coins.String())
	}
	if !in.Coins.IsPositive() {
		return sdk.ErrInvalidCoins(in.Coins.String())
	}
	return nil
}

// NewInput - create a transaction input, used with MsgSend
func NewInput(addr sdk.AccAddress, coins sdk.Coins) Input {
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
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
}

// Return bytes to sign for Output
func (out Output) GetSignBytes() []byte {
	bin, err := msgCdc.MarshalJSON(out)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bin)
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() sdk.Error {
	if len(out.Address) == 0 {
		return sdk.ErrInvalidAddress(out.Address.String())
	}
	if !out.Coins.IsValid() {
		return sdk.ErrInvalidCoins(out.Coins.String())
	}
	if !out.Coins.IsPositive() {
		return sdk.ErrInvalidCoins(out.Coins.String())
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgSend
func NewOutput(addr sdk.AccAddress, coins sdk.Coins) Output {
	output := Output{
		Address: addr,
		Coins:   coins,
	}
	return output
}
