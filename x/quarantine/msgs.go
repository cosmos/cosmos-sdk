package quarantine

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	qerrors "github.com/cosmos/cosmos-sdk/x/quarantine/errors"
)

var _ sdk.Msg = &MsgOptIn{}

// NewMsgOptIn creates a new msg to opt in to account quarantine.
func NewMsgOptIn(toAddr sdk.AccAddress) *MsgOptIn {
	return &MsgOptIn{
		ToAddress: toAddr.String(),
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgOptIn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgOptIn) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgOptOut{}

// NewMsgOptOut creates a new msg to opt out of account quarantine.
func NewMsgOptOut(toAddr sdk.AccAddress) *MsgOptOut {
	return &MsgOptOut{
		ToAddress: toAddr.String(),
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgOptOut) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgOptOut) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgAccept{}

// NewMsgAccept creates a new msg to accept quarantined funds.
func NewMsgAccept(toAddr sdk.AccAddress, fromAddrsStrs []string, permanent bool) *MsgAccept {
	return &MsgAccept{
		ToAddress:     toAddr.String(),
		FromAddresses: fromAddrsStrs,
		Permanent:     permanent,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgAccept) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if len(msg.FromAddresses) == 0 {
		return sdkerrors.ErrUnknownAddress.Wrap("at least one from address is required")
	}
	for i, addr := range msg.FromAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address[%d]: %s", i, err)
		}
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgAccept) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgDecline{}

// NewMsgDecline creates a new msg to decline quarantined funds.
func NewMsgDecline(toAddr sdk.AccAddress, fromAddrsStrs []string, permanent bool) *MsgDecline {
	return &MsgDecline{
		ToAddress:     toAddr.String(),
		FromAddresses: fromAddrsStrs,
		Permanent:     permanent,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgDecline) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if len(msg.FromAddresses) == 0 {
		return sdkerrors.ErrUnknownAddress.Wrap("at least one from address is required")
	}
	for i, addr := range msg.FromAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid from address[%d]: %s", i, err)
		}
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgDecline) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}

var _ sdk.Msg = &MsgUpdateAutoResponses{}

// NewMsgUpdateAutoResponses creates a new msg to update quarantined auto-responses.
func NewMsgUpdateAutoResponses(toAddr sdk.AccAddress, updates []*AutoResponseUpdate) *MsgUpdateAutoResponses {
	return &MsgUpdateAutoResponses{
		ToAddress: toAddr.String(),
		Updates:   updates,
	}
}

// ValidateBasic does simple stateless validation of this Msg.
func (msg MsgUpdateAutoResponses) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.ToAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid to address: %s", err)
	}
	if len(msg.Updates) == 0 {
		return qerrors.ErrInvalidValue.Wrap("no updates")
	}
	for i, update := range msg.Updates {
		if err := update.Validate(); err != nil {
			return errors.Wrapf(err, "invalid update %d", i+1)
		}
	}
	return nil
}

// GetSigners returns the addresses of required signers of this Msg.
func (msg MsgUpdateAutoResponses) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.ToAddress)
	return []sdk.AccAddress{addr}
}
