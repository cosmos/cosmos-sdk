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
	msg := MsgChangePubKey{
		Address: address,
	}
	msg.SetPubKey(pubKey)
	return &msg
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

	if len(msg.PubKey) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "pubkey should not be empty")
	}

	if len(msg.GetPubKey().Address()) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "pubkey should be able to associated to a valid address")
	}

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

// GetPubKey returns public key
func (msg MsgChangePubKey) GetPubKey() (pk crypto.PubKey) {
	if len(msg.PubKey) == 0 {
		return nil
	}

	amino.MustUnmarshalBinaryBare(msg.PubKey, &pk)
	return pk
}

// SetPubKey set public key
func (msg *MsgChangePubKey) SetPubKey(pubKey crypto.PubKey) error {
	if pubKey == nil {
		msg.PubKey = nil
	} else {
		msg.PubKey = amino.MustMarshalBinaryBare(pubKey)
	}

	return nil
}
