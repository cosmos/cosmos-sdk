package cool

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// a really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type MsgSetTrend struct {
	Sender sdk.AccAddress
	Cool   string
}

// genesis state - specify genesis trend
type Genesis struct {
	Trend string `json:"trend"`
}

// new cool message
func NewMsgSetTrend(sender sdk.AccAddress, cool string) MsgSetTrend {
	return MsgSetTrend{
		Sender: sender,
		Cool:   cool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgSetTrend{}

// nolint
func (msg MsgSetTrend) Type() string                 { return "cool" }
func (msg MsgSetTrend) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }
func (msg MsgSetTrend) String() string {
	return fmt.Sprintf("MsgSetTrend{Sender: %v, Cool: %v}", msg.Sender, msg.Cool)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgSetTrend) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).TraceSDK("")
	}
	if strings.Contains(msg.Cool, "hot") {
		return sdk.ErrUnauthorized("").TraceSDK("hot is not cool")
	}
	if strings.Contains(msg.Cool, "warm") {
		return sdk.ErrUnauthorized("").TraceSDK("warm is not very cool")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg MsgSetTrend) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

//_______________________________________________________________________

// A message type to quiz how cool you are. these fields are can be entirely
// arbitrary and custom to your message
type MsgQuiz struct {
	Sender     sdk.AccAddress
	CoolAnswer string
}

// New cool message
func NewMsgQuiz(sender sdk.AccAddress, coolerthancool string) MsgQuiz {
	return MsgQuiz{
		Sender:     sender,
		CoolAnswer: coolerthancool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgQuiz{}

// nolint
func (msg MsgQuiz) Type() string                 { return "cool" }
func (msg MsgQuiz) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }
func (msg MsgQuiz) String() string {
	return fmt.Sprintf("MsgQuiz{Sender: %v, CoolAnswer: %v}", msg.Sender, msg.CoolAnswer)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgQuiz) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).TraceSDK("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg MsgQuiz) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
