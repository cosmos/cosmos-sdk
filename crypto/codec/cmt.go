package codec

import (
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/encoding"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FromTmProtoPublicKey converts a TM's cmtprotocrypto.PublicKey into our own PubKey.
func FromTmProtoPublicKey(protoPk cmtprotocrypto.PublicKey) (cryptotypes.PubKey, error) {
	switch protoPk := protoPk.Sum.(type) {
	case *cmtprotocrypto.PublicKey_Ed25519:
		return &ed25519.PubKey{
			Key: protoPk.Ed25519,
		}, nil
	case *cmtprotocrypto.PublicKey_Secp256K1:
		return &secp256k1.PubKey{
			Key: protoPk.Secp256K1,
		}, nil
	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v from Tendermint public key", protoPk)
	}
}

// ToTmProtoPublicKey converts our own PubKey to TM's cmtprotocrypto.PublicKey.
func ToTmProtoPublicKey(pk cryptotypes.PubKey) (cmtprotocrypto.PublicKey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return cmtprotocrypto.PublicKey{
			Sum: &cmtprotocrypto.PublicKey_Ed25519{
				Ed25519: pk.Key,
			},
		}, nil
	case *secp256k1.PubKey:
		return cmtprotocrypto.PublicKey{
			Sum: &cmtprotocrypto.PublicKey_Secp256K1{
				Secp256K1: pk.Key,
			},
		}, nil
	default:
		return cmtprotocrypto.PublicKey{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to Tendermint public key", pk)
	}
}

// FromTmPubKeyInterface converts TM's cmtcrypto.PubKey to our own PubKey.
func FromTmPubKeyInterface(tmPk cmtcrypto.PubKey) (cryptotypes.PubKey, error) {
	tmProtoPk, err := encoding.PubKeyToProto(tmPk)
	if err != nil {
		return nil, err
	}

	return FromTmProtoPublicKey(tmProtoPk)
}

// ToTmPubKeyInterface converts our own PubKey to TM's cmtcrypto.PubKey.
func ToTmPubKeyInterface(pk cryptotypes.PubKey) (cmtcrypto.PubKey, error) {
	tmProtoPk, err := ToTmProtoPublicKey(pk)
	if err != nil {
		return nil, err
	}

	return encoding.PubKeyFromProto(tmProtoPk)
}
