package cool

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type SetCoolMsg struct {
	Sender sdk.Address
	Cool   string
}

// New cool message
func NewSetCoolMsg(sender sdk.Address, cool string) SetCoolMsg {
	return SetCoolMsg{
		Sender: sender,
		Cool:   cool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = SetCoolMsg{}

// nolint
func (msg SetCoolMsg) Type() string                            { return "cool" }
func (msg SetCoolMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg SetCoolMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg SetCoolMsg) String() string {
	return fmt.Sprintf("SetCoolMsg{Sender: %v, Cool: %v}", msg.Sender, msg.Cool)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg SetCoolMsg) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnrecognizedAddress(msg.Sender).Trace("")
	}
	if strings.Contains(msg.Cool, "hot") {
		return sdk.ErrUnauthorized("").Trace("hot is not cool")
	}
	if strings.Contains(msg.Cool, "warm") {
		return sdk.ErrUnauthorized("").Trace("warm is not very cool")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg SetCoolMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//_______________________________________________________________________

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type TestYourCoolnessMsg struct {
	Sender         sdk.Address
	CoolerThanCool string
}

// New cool message
func NewTestYourCoolnessMsg(sender sdk.Address, coolerthancool string) TestYourCoolnessMsg {
	return TestYourCoolnessMsg{
		Sender:         sender,
		CoolerThanCool: coolerthancool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = TestYourCoolnessMsg{}

// nolint
func (msg TestYourCoolnessMsg) Type() string                            { return "cool" }
func (msg TestYourCoolnessMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg TestYourCoolnessMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg TestYourCoolnessMsg) String() string {
	return fmt.Sprintf("TestYourCoolnessMsg{Sender: %v, CoolerThanCool: %v}", msg.Sender, msg.CoolerThanCool)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg TestYourCoolnessMsg) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnrecognizedAddress(msg.Sender).Trace("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg TestYourCoolnessMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
