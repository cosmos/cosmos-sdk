package std

import (
	"errors"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

//TODO: Deprecate this file when tendermint 0.34 is released

// PubKeyToProto takes crypto.PubKey and transforms it to a protobuf Pubkey
func PubKeyToProto(k crypto.PubKey) (PublicKey, error) {
	var kp PublicKey
	switch k := k.(type) {
	case sr25519.PubKeySr25519:

		kp = PublicKey{
			Sum: &PublicKey_Sr25519{
				Sr25519: k[:],
			},
		}
	case ed25519.PubKeyEd25519:
		kp = PublicKey{
			Sum: &PublicKey_Ed25519{
				Ed25519: k[:],
			},
		}
	case secp256k1.PubKeySecp256k1:
		kp = PublicKey{
			Sum: &PublicKey_Secp256K1{
				Secp256K1: k[:],
			},
		}
	default:
		return kp, errors.New("toproto: key type is not supported")
	}
	return kp, nil
}

// PubKeyFromProto takes a protobuf Pubkey and transforms it to a crypto.Pubkey
func PubKeyFromProto(k PublicKey) (crypto.PubKey, error) {
	switch k := k.Sum.(type) {
	case *PublicKey_Ed25519:
		var pk ed25519.PubKeyEd25519
		copy(pk[:], k.Ed25519)
		return pk, nil
	case *PublicKey_Sr25519:
		var pk sr25519.PubKeySr25519
		copy(pk[:], k.Sr25519)
		return pk, nil
	case *PublicKey_Secp256K1:
		var pk secp256k1.PubKeySecp256k1
		copy(pk[:], k.Secp256K1)
		return pk, nil
	default:
		return nil, errors.New("fromproto: key type not supported")
	}
}

// PrivKeyToProto takes crypto.PrivKey and transforms it to a protobuf PrivKey
func PrivKeyToProto(k crypto.PrivKey) (PrivateKey, error) {
	var kp PrivateKey
	switch k := k.(type) {
	case ed25519.PrivKeyEd25519:
		kp = PrivateKey{
			Sum: &PrivateKey_Ed25519{
				Ed25519: k[:],
			},
		}
	case sr25519.PrivKeySr25519:
		kp = PrivateKey{
			Sum: &PrivateKey_Sr25519{
				Sr25519: k[:],
			},
		}
	case secp256k1.PrivKeySecp256k1:
		kp = PrivateKey{
			Sum: &PrivateKey_Secp256K1{
				Secp256K1: k[:],
			},
		}
	default:
		return kp, errors.New("toproto: key type is not supported")
	}
	return kp, nil
}

// PrivKeyFromProto takes a protobuf PrivateKey and transforms it to a crypto.PrivKey
func PrivKeyFromProto(k PrivateKey) (crypto.PrivKey, error) {
	switch k := k.Sum.(type) {
	case *PrivateKey_Ed25519:
		var pk ed25519.PrivKeyEd25519
		copy(pk[:], k.Ed25519)
		return pk, nil
	case *PrivateKey_Sr25519:
		var pk sr25519.PrivKeySr25519
		copy(pk[:], k.Sr25519)
		return pk, nil
	case *PrivateKey_Secp256K1:
		var pk secp256k1.PrivKeySecp256k1
		copy(pk[:], k.Secp256K1)
		return pk, nil
	default:
		return nil, errors.New("fromproto: key type not supported")
	}
}
