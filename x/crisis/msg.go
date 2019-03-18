package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgVerifyInvariance - message struct to verify a particular invariance
type MsgVerifyInvariance struct {
	Sender          sdk.AccAddress `json:"sender"`
	InvarianceRoute string         `json:"invariance_route"`
}

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgVerifyInvariance{}

// MsgVerifyInvariance - create a new MsgVerifyInvariance object
func NewMsgVerifyInvariance(sender sdk.AccAddress, invarianceRoute string) MsgVerifyInvariance {
	return MsgVerifyInvariance{
		Sender:          sender,
		InvarianceRoute: invarianceRoute,
	}
}

//nolint
func (msg MsgVerifyInvariance) Route() string { return ModuleName }
func (msg MsgVerifyInvariance) Type() string  { return "verify_invariance" }

// get the bytes for the message signer to sign on
func (msg MsgVerifyInvariance) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the sign bytes for the msg MsgVerifyInvariance
func (msg MsgVerifyInvariance) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVerifyInvariance) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return ErrNilSender(DefaultCodespace)
	}
	return nil
}
