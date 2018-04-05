package cool

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A really cool msg type, these fields are can be entirely arbitrary and
// custom to your message
type SetTrendMsg struct {
	Sender sdk.Address
	Cool   string
}

// Genesis state - specify genesis trend
type CoolGenesis struct {
	Trend string `json:"trend"`
}

// New cool message
func NewSetTrendMsg(sender sdk.Address, cool string) SetTrendMsg {
	return SetTrendMsg{
		Sender: sender,
		Cool:   cool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = SetTrendMsg{}

// nolint
func (msg SetTrendMsg) Type() string                            { return "cool" }
func (msg SetTrendMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg SetTrendMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg SetTrendMsg) String() string {
	return fmt.Sprintf("SetTrendMsg{Sender: %v, Cool: %v}", msg.Sender, msg.Cool)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg SetTrendMsg) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).Trace("")
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
func (msg SetTrendMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

//_______________________________________________________________________

// A message type to quiz how cool you are. these fields are can be entirely
// arbitrary and custom to your message
type QuizMsg struct {
	Sender     sdk.Address
	CoolAnswer string
}

// New cool message
func NewQuizMsg(sender sdk.Address, coolerthancool string) QuizMsg {
	return QuizMsg{
		Sender:     sender,
		CoolAnswer: coolerthancool,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = QuizMsg{}

// nolint
func (msg QuizMsg) Type() string                            { return "cool" }
func (msg QuizMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg QuizMsg) GetSigners() []sdk.Address               { return []sdk.Address{msg.Sender} }
func (msg QuizMsg) String() string {
	return fmt.Sprintf("QuizMsg{Sender: %v, CoolAnswer: %v}", msg.Sender, msg.CoolAnswer)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg QuizMsg) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).Trace("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg QuizMsg) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}
