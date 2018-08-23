package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc = wire.NewCodec()

// name to identify transaction types
const MsgType = "slashing"

// verify interface at compile time
var _ sdk.Msg = &MsgUnjail{}

// MsgUnjail - struct for unjailing jailed validator
type MsgUnjail struct {
	ValidatorAddr sdk.AccAddress `json:"address"` // address of the validator owner
}

func NewMsgUnjail(validatorAddr sdk.AccAddress) MsgUnjail {
	return MsgUnjail{
		ValidatorAddr: validatorAddr,
	}
}

//nolint
func (msg MsgUnjail) Type() string                 { return MsgType }
func (msg MsgUnjail) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.ValidatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgUnjail) GetSignBytes() []byte {
	b, err := cdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgUnjail) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	return nil
}
