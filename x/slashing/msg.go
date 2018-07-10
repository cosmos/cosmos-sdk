package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc = wire.NewCodec()

// name to identify transaction types
const MsgType = "slashing"

// verify interface at compile time
var _ sdk.Msg = &MsgUnrevoke{}

// MsgUnrevoke - struct for unrevoking revoked validator
type MsgUnrevoke struct {
	ValidatorAddr sdk.AccAddress `json:"address"` // address of the validator owner
}

func NewMsgUnrevoke(validatorAddr sdk.AccAddress) MsgUnrevoke {
	return MsgUnrevoke{
		ValidatorAddr: validatorAddr,
	}
}

//nolint
func (msg MsgUnrevoke) Type() string                 { return MsgType }
func (msg MsgUnrevoke) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.ValidatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgUnrevoke) GetSignBytes() []byte {
	b, err := cdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// quick validity check
func (msg MsgUnrevoke) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	return nil
}
