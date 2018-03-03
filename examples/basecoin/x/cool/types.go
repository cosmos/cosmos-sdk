package cool

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type SetWhatCoolMsg struct {
	Sender   sdk.Address
	WhatCool string
}

// New cool message
func NewSetWhatCoolMsg(sender sdk.Address, whatcool string) SetWhatCoolMsg {
	return SetWhatCoolMsg{
		Sender:   sender,
		WhatCool: whatcool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = SetWhatCoolMsg{}

// nolint
func (msg SetWhatCoolMsg) Type() string                            { return "cool" }
func (msg SetWhatCoolMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg SetWhatCoolMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg SetWhatCoolMsg) String() string {
	return fmt.Sprintf("SetWhatCoolMsg{Sender: %v, WhatCool: %v}", msg.Sender, msg.WhatCool)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg SetWhatCoolMsg) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnrecognizedAddress(msg.Sender).Trace("")
	}
	if strings.Contains(msg.WhatCool, "hot") {
		return sdk.ErrUnauthorized("").Trace("hot is not cool")
	}
	if strings.Contains(msg.WhatCool, "warm") {
		return sdk.ErrUnauthorized("").Trace("warm is not very cool")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg SetWhatCoolMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//_______________________________________________________________________

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type WhatCoolMsg struct {
	Sender         sdk.Address
	CoolerThanCool string
}

// New cool message
func NewWhatCoolMsg(sender sdk.Address, coolerthancool string) WhatCoolMsg {
	return WhatCoolMsg{
		Sender:         sender,
		CoolerThanCool: coolerthancool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = WhatCoolMsg{}

// nolint
func (msg WhatCoolMsg) Type() string                            { return "cool" }
func (msg WhatCoolMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg WhatCoolMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg WhatCoolMsg) String() string {
	return fmt.Sprintf("WhatCoolMsg{Sender: %v, CoolerThanCool: %v}", msg.Sender, msg.CoolerThanCool)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg WhatCoolMsg) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnrecognizedAddress(msg.Sender).Trace("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg WhatCoolMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
