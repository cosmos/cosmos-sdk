package codec

import (
	"cosmossdk.io/errors"

	bls12_381 "github.com/cosmos/cosmos-sdk/crypto/keys/bls12_381"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func PubKeyToProto(pk cryptotypes.JSONPubKey) (cryptotypes.PubKey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return &ed25519.PubKey{
			Key: pk.Bytes(),
		}, nil
	case *secp256k1.PubKey:
		return &secp256k1.PubKey{
			Key: pk.Bytes(),
		}, nil
	case *bls12_381.PubKey:
		return &bls12_381.PubKey{
			Key: pk.Bytes(),
		}, nil
	default:
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to proto public key", pk)
	}
}

func PubKeyFromProto(pk cryptotypes.PubKey) (cryptotypes.JSONPubKey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return &ed25519.PubKey{
			Key: pk.Bytes(),
		}, nil
	case *secp256k1.PubKey:
		return &secp256k1.PubKey{
			Key: pk.Bytes(),
		}, nil
	case *bls12_381.PubKey:
		return &bls12_381.PubKey{
			Key: pk.Bytes(),
		}, nil
	default:
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v from proto public key", pk)
	}
}
