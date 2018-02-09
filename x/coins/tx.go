package coins

import (
	"encoding/json"
	"fmt"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SendMsg - high level transaction of the coin module
type SendMsg struct {
	FromAddress  crypto.Address `json:"address"`
	ToAddress    crypto.Address `json:"address"`
	Coins        Coins          `json:"coins"`
}

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

	if len(msg.FromAddress) == 0 {
		return ErrInvalidAddress(msg.FromAddress.String())
	}
	if len(msg.ToAddress) == 0 {
		return ErrInvalidAddress(msg.ToAddress.String())
	}
	if !msg.Coins.IsValid() {
		return ErrInvalidCoins(msg.Coins.String())
	}
	if !msg.Coins.IsPositive() {
		return ErrInvalidCoins(msg.Coins.String())
	}
	return nil
}

func (msg SendMsg) String() string {
	return fmt.Sprintf("SendMsg{%v->%v:%v}", msg.FromAddress, msg.ToAddress, msg.Coins)
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
	return []crypto.Address{msg.FromAddress}
}

//----------------------------------------
// IssueMsg

// IssueMsg - high level transaction of the coin module
type IssueMsg struct {
	Banker       crypto.Address `json:"banker"`
	ToAddress    crypto.Address `json:"address"`
	Coins        Coins          `json:"coins"`
}

// NewIssueMsg - construct arbitrary multi-in, multi-out send msg.
func NewIssueMsg(banker crypto.Address, toAddress crypto.Address, coins Coins) IssueMsg {
	return IssueMsg{Banker: banker, ToAddress: toAddress, Coins: coins}
}

// Implements Msg.
func (msg IssueMsg) Type() string { return "bank" } // TODO: "bank/send"

// Implements Msg.
func (msg IssueMsg) ValidateBasic() sdk.Error {
	// XXX
	if len(msg.Banker) == 0 {
		return ErrInvalidAddress(msg.Banker.String())
	}
	if len(msg.ToAddress) == 0 {
		return ErrInvalidAddress(msg.ToAddress.String())
	}
	if !msg.Coins.IsValid() {
		return ErrInvalidCoins(msg.Coins.String())
	}
	if !msg.Coins.IsPositive() {
		return ErrInvalidCoins(msg.Coins.String())
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
func (msg IssueMsg) GetSigners() []crypto.Address {
	return []crypto.Address{msg.Banker}
}
