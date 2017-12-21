package coin

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tmlibs/common"
)

// NOTE: How the app decodes a SendMsg or IssueMsg is up to
// the app implementation.  Do not include parsing logic
// here.
type CoinMsg interface {
	AssertIsCoinMsg()
}

func (_ SendMsg) AssertIsCoinMsg()  {}
func (_ IssueMsg) AssertIsCoinMsg() {}

//----------------------------------------
// SendMsg

// SendMsg - high level transaction of the coin module
type SendMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

var _ CoinMsg = SendMsg

func NewSendMsg(in []Input, out []Output) SendMsg {
	return SendMsg{Inputs: in, Outputs: out}
}

// Implements types.Msg.
func (msg SendMsg) Get(key interface{}) (value interface{}) {
	panic("not implemented yet") // XXX
}

// Implements types.Msg.
func (msg SendMsg) SignBytes() []byte {
	panic("SendMsg does not implement SignBytes. Implement it by embedding SendMsg in a custom struct")
}

// Implements types.Msg.
func (msg SendMsg) ValidateBasic() error {
	return nil
}

// Implements types.Msg.
func (msg SendMsg) ValidateBasic() error {
	if len(msg.Inputs) == 0 {
		return ErrInvalidInput("SendMsg needs 1 or more inputs")
	}
	if len(msg.Outputs) == 0 {
		return ErrInvalidOutput("SendMsg needs 1 or more outputs")
	}

	// While tallying totals, validate inputs and outputs.
	var totalIn, totalOut Coins
	for _, in := range msg.Inputs {
		if err := in.ValidateBasic(); err != nil {
			return ErrInvalidInput("").WithCause(err)
		}
		totalIn = totalIn.Plus(in.Coins)
	}
	for _, out := range msg.Outputs {
		if err := out.ValidateBasic(); err != nil {
			return ErrInvalidOutput("").WithCause(err)
		}
		totalOut = totalOut.Plus(out.Coins)
	}

	// Ensure that totals match.
	// TODO: Handle fees, whether too low, legative, or too high.
	if !totalIn.IsEqual(totalOut) {
		return ErrInsufficientFunds("")
	}

	// All good!
	return nil
}

// Implements types.Msg.
func (msg SendMsg) Signers() [][]byte {
	panic("not implemented yet") // XXX
}

func (msg SendMsg) String() string {
	return fmt.Sprintf("SendMsg{%v->%v}", msg.Inputs, msg.Outputs)
}

//----------------------------------------
// IssueMsg

// IssueMsg allows issuer to issue new coins.
type IssueMsg struct {
	Issuer []byte `json:"issuer"`
	Target []byte `json:"target"`
	Coins  `json:"coins"`
}

var _ CoinMsg = IssueMsg

func NewIssueMsg(issuer []byte, target []byte, coins Coins) IssueMsg {
	return IssueMsg{
		Issuer: issuer,
		Target: target,
		Coins:  coins,
	}
}

// Implements types.Msg.
func (msg IssueMsg) Get(key interface{}) (value interface{}) {
	panic("not implemented yet") // XXX
}

// Implements types.Msg.
func (msg IssueMsg) SignBytes() []byte {
	panic("IssueMsg does not implement SignBytes. Implement it by embedding IssueMsg in a custom struct")
}

// Implements types.Msg.
func (msg IssueMsg) ValidateBasic() error {
	return nil
}

// Implements types.Msg.
func (msg IssueMsg) Signers() [][]byte {
	panic("not implemented yet") // XXX
}

func (msg IssueMsg) String() string {
	return fmt.Sprintf("IssueMsg{%X:%v->%X}",
		msg.Issuer, msg.Coins, msg.Target)
}

//----------------------------------------
// Input

// Input is a source of coins in a transaction.
type Input struct {
	Address cmn.Bytes `json:"address"`
	Coins   Coins     `json:"coins"`
}

func NewInput(addr []byte, coins Coins) Input {
	return Input{
		Address: addr,
		Coins:   coins,
	}
}

func (inp Input) ValidateBasic() error {
	if !auth.IsValidAddress(inp.Address) {
		return ErrInvalidAddress(fmt.Sprintf(
			"Invalid input address %X", inp.Address))
	}
	if !inp.Coins.IsValid() {
		return ErrInvalidInput(fmt.Sprintf(
			"Input coins not valid: %v", inp.Coins))
	}
	if !inp.Coins.IsPositive() {
		return ErrInvalidInput("Input coins must be positive")
	}
	return nil
}

func (inp Input) String() string {
	return fmt.Sprintf("Input{%v,%v}", inp.Address, inp.Coins)
}

//----------------------------------------
// Output

// Output is a destination of coins in a transaction.
type Output struct {
	Address cmn.Bytes `json:"address"`
	Coins   Coins     `json:"coins"`
}

func NewOutput(addr []byte, coins Coins) Output {
	output := Output{
		Address: addr,
		Coins:   coins,
	}
	return output
}

func (out Output) ValidateBasic() error {
	if !auth.IsValidAddress(out.Address) {
		return ErrInvalidAddress(fmt.Sprintf(
			"Invalid output address %X", out.Address))
	}
	if !out.Coins.IsValid() {
		return ErrInvalidOutput(fmt.Sprintf(
			"Output coins not valid: %v", out.Coins))
	}
	if !out.Coins.IsPositive() {
		return ErrInvalidOutput("Output coins must be positive")
	}
	return nil
}

func (out Output) String() string {
	return fmt.Sprintf("Output{%X,%v}", out.Address, out.Coins)
}
