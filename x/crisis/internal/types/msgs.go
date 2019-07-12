package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgVerifyInvariant - message struct to verify a particular invariance
type MsgVerifyInvariant struct {
	Sender              sdk.AccAddress `json:"sender" yaml:"sender"`
	InvariantModuleName string         `json:"invariant_module_name" yaml:"invariant_module_name"`
	InvariantRoute      string         `json:"invariant_route" yaml:"invariant_route"`
}

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgVerifyInvariant{}

// NewMsgVerifyInvariant creates a new MsgVerifyInvariant object
func NewMsgVerifyInvariant(sender sdk.AccAddress, invariantModuleName,
	invariantRoute string) MsgVerifyInvariant {

	return MsgVerifyInvariant{
		Sender:              sender,
		InvariantModuleName: invariantModuleName,
		InvariantRoute:      invariantRoute,
	}
}

//nolint
func (msg MsgVerifyInvariant) Route() string { return ModuleName }
func (msg MsgVerifyInvariant) Type() string  { return "verify_invariant" }

// get the bytes for the message signer to sign on
func (msg MsgVerifyInvariant) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the sign bytes for the msg MsgVerifyInvariant
func (msg MsgVerifyInvariant) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVerifyInvariant) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return ErrNilSender(DefaultCodespace)
	}
	return nil
}

// FullInvariantRoute - get the messages full invariant route
func (msg MsgVerifyInvariant) FullInvariantRoute() string {
	return msg.InvariantModuleName + "/" + msg.InvariantRoute
}
