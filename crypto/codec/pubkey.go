package codec

import (
	"github.com/cometbft/cometbft/crypto/bls12381"

	"cosmossdk.io/errors"

	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12_381"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// PubKeyToProto converts a JSON public key (in `cryptokeys.JSONPubkey` format) to its corresponding protobuf public key type.
//
// Parameters:
// - pk: A `cryptokeys.JSONPubkey` containing the public key and its type.
//
// Returns:
// - cryptotypes.PubKey: The protobuf public key corresponding to the provided JSON public key.
// - error: An error if the key type is invalid or unsupported.
func PubKeyToProto(pk cryptokeys.JSONPubkey) (cryptotypes.PubKey, error) {
	switch pk.KeyType {
	case ed25519.PubKeyName:
		return &ed25519.PubKey{
			Key: pk.Value,
		}, nil
	case secp256k1.PubKeyName:
		return &secp256k1.PubKey{
			Key: pk.Value,
		}, nil
	case bls12381.PubKeyName:
		return &bls12_381.PubKey{
			Key: pk.Value,
		}, nil
	default:
		return nil, errors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to proto public key", pk)
	}
}

// PubKeyFromProto converts a protobuf public key (in `cryptotypes.PubKey` format) to a JSON public key format (`cryptokeys.JSONPubkey`).
//
// Parameters:
// - pk: A `cryptotypes.PubKey` which is the protobuf representation of a public key.
//
// Returns:
// - cryptokeys.JSONPubkey: The JSON-formatted public key corresponding to the provided protobuf public key.
// - error: An error if the key type is invalid or unsupported.
func PubKeyFromProto(pk cryptotypes.PubKey) (cryptokeys.JSONPubkey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return cryptokeys.JSONPubkey{
			KeyType: ed25519.PubKeyName,
			Value:   pk.Bytes(),
		}, nil
	case *secp256k1.PubKey:
		return cryptokeys.JSONPubkey{
			KeyType: secp256k1.PubKeyName,
			Value:   pk.Bytes(),
		}, nil
	case *bls12_381.PubKey:
		return cryptokeys.JSONPubkey{
			KeyType: bls12381.PubKeyName,
			Value:   pk.Bytes(),
		}, nil
	default:
		return cryptokeys.JSONPubkey{}, errors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v from proto public key", pk)
	}
}
