package bank

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SendMsg - high level transaction of the coin module
type SendMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

var _ sdk.Msg = SendMsg{}

// NewSendMsg - construct arbitrary multi-in, multi-out send msg.
func NewSendMsg(in []Input, out []Output) SendMsg {
	return SendMsg{Inputs: in, Outputs: out}
}

// Implements Msg.
func (msg SendMsg) Type() string { return "bank" } // TODO: "bank/send"

// Implements Msg.
func (msg SendMsg) ValidateBasic() sdk.Error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs().Trace("")
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs().Trace("")
	}
	// make sure all inputs and outputs are individually valid
	var totalIn, totalOut sdk.Coins
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
		return sdk.ErrInvalidCoins(totalIn.String()).Trace("inputs and outputs don't match")
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
func (msg SendMsg) GetSigners() []sdk.Address {
	addrs := make([]sdk.Address, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = in.Address
	}
	return addrs
}

//----------------------------------------
// IssueMsg

// IssueMsg - high level transaction of the coin module
type IssueMsg struct {
	Banker  sdk.Address `json:"banker"`
	Outputs []Output    `json:"outputs"`
}

// NewIssueMsg - construct arbitrary multi-in, multi-out send msg.
func NewIssueMsg(banker sdk.Address, out []Output) IssueMsg {
	return IssueMsg{Banker: banker, Outputs: out}
}

// Implements Msg.
func (msg IssueMsg) Type() string { return "bank" } // TODO: "bank/issue"

// Implements Msg.
func (msg IssueMsg) ValidateBasic() sdk.Error {
	// XXX
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs().Trace("")
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.Trace("")
		}
	}
	return nil
}

func (msg IssueMsg) String() string {
	return fmt.Sprintf("IssueMsg{%v#%v}", msg.Banker, msg.Outputs)
}

// Implements Msg.
func (msg IssueMsg) Get(key interface{}) (value interface{}) {
	return nil
}

// Implements Msg.
func (msg IssueMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg IssueMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Banker}
}

//----------------------------------------
// Input

// Transaction Output
type Input struct {
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
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

func (in Input) String() string {
	return fmt.Sprintf("Input{%v,%v}", in.Address, in.Coins)
}

// NewInput - create a transaction input, used with SendMsg
func NewInput(addr sdk.Address, coins sdk.Coins) Input {
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
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
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

func (out Output) String() string {
	return fmt.Sprintf("Output{%v,%v}", out.Address, out.Coins)
}

// NewOutput - create a transaction output, used with SendMsg
func NewOutput(addr sdk.Address, coins sdk.Coins) Output {
	output := Output{
		Address: addr,
		Coins:   coins,
	}
	return output
}
