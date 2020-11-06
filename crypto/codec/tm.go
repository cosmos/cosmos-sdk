package codec

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/encoding"
	tmprotocrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FromTmPublicKey converts a TM's tmprotocrypto.PublicKey into our own PubKey.
func FromTmPublicKey(protoPk tmprotocrypto.PublicKey) (cryptotypes.PubKey, error) {
	switch protoPk := protoPk.Sum.(type) {
	case *tmprotocrypto.PublicKey_Ed25519:
		return &ed25519.PubKey{
			Key: protoPk.Ed25519,
		}, nil
	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v from Tendermint public key", protoPk)
	}
}

// ToTmPublicKey converts our own PubKey to TM's tmprotocrypto.PublicKey.
func ToTmPublicKey(pk cryptotypes.PubKey) (tmprotocrypto.PublicKey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return tmprotocrypto.PublicKey{
			Sum: &tmprotocrypto.PublicKey_Ed25519{
				Ed25519: pk.Key,
			},
		}, nil
	default:
		return tmprotocrypto.PublicKey{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to Tendermint public key", pk)
	}
}

// FromTmPubKey converts TM's tmcrypto.PubKey to our own PubKey.
func FromTmPubKey(tmPk tmcrypto.PubKey) (cryptotypes.PubKey, error) {
	tmProtoPk, err := encoding.PubKeyToProto(tmPk)
	if err != nil {
		return nil, err
	}

	return FromTmPublicKey(tmProtoPk)
}

// ToTmPubKey converts our own PubKey to TM's tmcrypto.PubKey.
func ToTmPubKey(pk cryptotypes.PubKey) (tmcrypto.PubKey, error) {
	tmProtoPk, err := ToTmPublicKey(pk)
	if err != nil {
		return nil, err
	}

	return encoding.PubKeyFromProto(tmProtoPk)
}
