package types

import (
	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TypeMsgChangePubKey defines the type value for a MsgChangePubKey.
const TypeMsgChangePubKey = "msg_change_pubkey"

var _ sdk.Msg = &MsgChangePubKey{}

// NewMsgChangePubKey returns a reference to a new MsgChangePubKey.
func NewMsgChangePubKey(address string, pubKey cryptotypes.PubKey) *MsgChangePubKey {
	msg := MsgChangePubKey{
		Address: address,
		PubKey:  pubKey.Bytes(),
	}

	return &msg
}

// Type returns the message type for a MsgChangePubKey.
func (msg MsgChangePubKey) Type() string { return TypeMsgChangePubKey }

// ValidateBasic Implements Msg.
func (msg MsgChangePubKey) ValidateBasic() error {
	accAddress, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return err
	}
	if err := sdk.VerifyAddressFormat(accAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	if len(msg.PubKey) == 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidPubKey, "pubkey should not be empty")
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
	accAddress, _ := sdk.AccAddressFromBech32(msg.Address)
	return []sdk.AccAddress{accAddress}
}

// GetPubKey returns public key
func (msg MsgChangePubKey) GetPubKey(codec codec.BinaryCodec) (pk cryptotypes.PubKey, err error) {
	if err := codec.UnmarshalInterface([]byte(msg.PubKey), &pk); err != nil {
		return nil, err
	}

	return pk, nil
}
