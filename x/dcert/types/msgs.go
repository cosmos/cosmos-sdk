package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure Msg interface compliance at compile time
var _ sdk.Msg = &MsgVoteCircuitBreak{}
var _ sdk.Msg = &MsgVoteChangeDesc{}
var _ sdk.Msg = &MsgVoteSuspendDCERTMember{}

//__________________________________________________________________________
// MsgVoteCircuitBreak - message struct to verify a particular invariance
type MsgVoteCircuitBreak struct {
	Sender     sdk.AccAddress `json:"sender" yaml:"sender"`
	BreakRoute string         `json:"break_route" yaml:"break_route"` // empty route breaks all
}

// NewMsgVoteCircuitBreak creates a new MsgVoteCircuitBreak object
func NewMsgVoteCircuitBreak(sender sdk.AccAddress, breakRoute string) MsgVoteCircuitBreak {
	return MsgVoteCircuitBreak{
		Sender:     sender,
		BreakRoute: breakRoute,
	}
}

//nolint
func (msg MsgVoteCircuitBreak) Route() string { return ModuleName }
func (msg MsgVoteCircuitBreak) Type() string  { return "verify_invariant" }

// get the bytes for the message signer to sign on
func (msg MsgVoteCircuitBreak) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the sign bytes for the msg MsgVoteCircuitBreak
func (msg MsgVoteCircuitBreak) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVoteCircuitBreak) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return ErrNilSender(DefaultCodespace)
	}
	return nil
}

//__________________________________________________________________________
// MsgVoteCircuitBreak - message struct to verify a particular invariance
type MsgVoteChangeDesc struct {
	Sender  sdk.AccAddress `json:"sender" yaml:"sender"`
	NewDesc string         `json:"new_desc" yaml:"new_desc"`
}

// NewMsgVoteChangeDesc creates a new MsgVoteChangeDesc object
func NewMsgVoteChangeDesc(sender sdk.AccAddress, newDesc string) MsgVoteChangeDesc {
	return MsgVoteChangeDesc{
		Sender:  sender,
		NewDesc: newDesc,
	}
}

//nolint
func (msg MsgVoteChangeDesc) Route() string { return ModuleName }
func (msg MsgVoteChangeDesc) Type() string  { return "verify_invariant" }

// get the bytes for the message signer to sign on
func (msg MsgVoteChangeDesc) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{msg.Sender} }

// GetSignBytes gets the sign bytes for the msg MsgVoteCircuitBreak
func (msg MsgVoteChangeDesc) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVoteChangeDesc) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return ErrNilSender(DefaultCodespace)
	}
	if msg.NewDesc == "" {
		return ErrNilDescription(DefaultCodespace)
	}
	return nil
}

//__________________________________________________________________________
// MsgVoteCircuitBreak - message struct to verify a particular invariance
type MsgVoteSuspendDCERTMember struct {
	Sender  sdk.AccAddress `json:"sender" yaml:"sender"`
	Suspend sdk.AccAddress `json:"suspend" yaml:"suspend"`
}

// NewMsgVoteSuspendDCERTMember creates a new MsgVoteSuspendDCERTMember object
func NewMsgVoteSuspendDCERTMember(sender sdk.AccAddress, suspend sdk.AccAddress) MsgVoteSuspendDCERTMember {
	return MsgVoteSuspendDCERTMember{
		Sender:  sender,
		Suspend: suspend,
	}
}

//nolint
func (msg MsgVoteSuspendDCERTMember) Route() string { return ModuleName }
func (msg MsgVoteSuspendDCERTMember) Type() string  { return "verify_invariant" }

// get the bytes for the message signer to sign on
func (msg MsgVoteSuspendDCERTMember) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// GetSignBytes gets the sign bytes for the msg MsgVoteSuspendDCERTMember
func (msg MsgVoteSuspendDCERTMember) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgVoteSuspendDCERTMember) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return ErrNilSender(DefaultCodespace)
	}
	if msg.Suspend.Empty() {
		return ErrNilSuspend(DefaultCodespace)
	}
	return nil
}
