package coolmodule

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type CoolMsg struct {
	Sender         sdk.Address
	Coolerthancool string
}

// New cool message
func NewCoolMsg(sender sdk.Address, coolerthancool string) CoolMsg {
	return CoolMsg{
		Sender:         sender,
		Coolerthancool: coolerthancool,
	}
}

// enforce the msg type at compile time
var _ CoolMsg = sdk.Msg

// nolint
func (msg CoolMsg) Type() string                            { return "cool" }
func (msg CoolMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg CoolMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg CoolMsg) String() string                          { return fmt.Sprintf("CoolMsg{%v}", msg) }

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg CoolMsg) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return ErrNoOutputs().Trace("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg CoolMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//_______________________________________________________________________

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type BlowDryMsg struct {
	Sender sdk.Address
	What   sdk.Coin
}

// New cool message
func NewBlowDryMsg(sender sdk.Address, what sdk.Coin) BlowDryMsg {
	return BlowDryMsg{
		Sender: sender,
		What:   what,
	}
}

// enforce the msg type at compile time
var _ BlowDryMsg = sdk.Msg

// nolint
func (msg BlowDryMsg) Type() string                            { return "cool" }
func (msg BlowDryMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg BlowDryMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg BlowDryMsg) String() string                          { return fmt.Sprintf("BlowDryMsg{%v}", msg) }

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg BlowDryMsg) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return ErrNoOutputs().Trace("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg BlowDryMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
