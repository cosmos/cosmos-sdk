package slashing

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// name to identify transaction types
const MsgType = "slashing"

// verify interface at compile time
var _ sdk.Msg = &MsgUnrevoke{}

// MsgUnrevoke - struct for unrevoking revoked validator
type MsgUnrevoke struct {
	ValidatorAddr sdk.Address `json:"address"`
}

func NewMsgUnrevoke(validatorAddr sdk.Address) MsgUnrevoke {
	return MsgUnrevoke{
		ValidatorAddr: validatorAddr,
	}
}

//nolint
func (msg MsgUnrevoke) Type() string              { return MsgType }
func (msg MsgUnrevoke) GetSigners() []sdk.Address { return []sdk.Address{msg.ValidatorAddr} }

// get the bytes for the message signer to sign on
func (msg MsgUnrevoke) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return b
}

// quick validity check
func (msg MsgUnrevoke) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr == nil {
		return ErrBadValidatorAddr(DefaultCodespace)
	}
	return nil
}
