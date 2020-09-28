package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto"
)

// TypeMsgChangePubKey defines the type value for a MsgChangePubKey.
const TypeMsgChangePubKey = "msg_change_pubkey"

var _ sdk.Msg = &MsgChangePubKey{}

// NewMsgChangePubKey returns a reference to a new MsgChangePubKey.
func NewMsgChangePubKey(address sdk.AccAddress, pubKey crypto.PubKey) *MsgChangePubKey {
	return &MsgChangePubKey{
		Address: address,
		PubKey:  pubKey,
	}
}

// Route returns the message route for a MsgChangePubKey.
func (msg MsgChangePubKey) Route() string { return RouterKey }

// Type returns the message type for a MsgChangePubKey.
func (msg MsgChangePubKey) Type() string { return TypeMsgChangePubKey }

// ValidateBasic Implements Msg.
func (msg MsgChangePubKey) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(msg.Address); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	// TODO should validate PubKey
	// problem: there can be several pubKey types

	return nil
}

// GetSignBytes returns the bytes all expected signers must sign over for a
// MsgChangePubKey.
func (msg MsgChangePubKey) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgChangePubKey.
func (msg MsgChangePubKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Address}
}
