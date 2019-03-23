package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgVerifyInvariant - message struct to verify a particular invariance
type MsgVerifyInvariant struct {
	Sender         sdk.AccAddress `json:"sender"`
	InvariantRoute string         `json:"invariant_route"`
}

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgVerifyInvariant{}

// MsgVerifyInvariant - create a new MsgVerifyInvariant object
func NewMsgVerifyInvariance(sender sdk.AccAddress, invariantRoute string) MsgVerifyInvariant {
	return MsgVerifyInvariant{
		Sender:         sender,
		InvariantRoute: invariantRoute,
	}
}

//nolint
func (msg MsgVerifyInvariant) Route() string { return ModuleName }
func (msg MsgVerifyInvariant) Type() string  { return "verify_invariant" }

// get the bytes for the message signer to sign on
func (msg MsgVerifyInvariant) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the sign bytes for the msg MsgVerifyInvariant
func (msg MsgVerifyInvariant) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVerifyInvariant) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return ErrNilSender(DefaultCodespace)
	}
	return nil
}
