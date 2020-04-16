package std

import (
	"errors"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
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
	case multisig.PubKeyMultisigThreshold:
		pk := make([]*PublicKey, len(k.PubKeys))
		for i := 0; i < len(k.PubKeys); i++ {
			pkp, err := PubKeyToProto(k.PubKeys[i])
			if err != nil {
				return PublicKey{}, err
			}
			pk[i] = &pkp
		}
		kp = PublicKey{
			Sum: &PublicKey_Multisig{
				Multisig: &PubKeyMultisigThreshold{
					K:       uint32(k.K),
					PubKeys: pk,
				},
			},
		}
	default:
		return kp, fmt.Errorf("toproto: key type %T is not supported", k)
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
	case *PublicKey_Multisig:
		pk := make([]crypto.PubKey, len(k.Multisig.PubKeys))
		for i := range k.Multisig.PubKeys {
			pkp, err := PubKeyFromProto(*k.Multisig.PubKeys[i])
			if err != nil {
				return nil, err
			}
			pk[i] = pkp
		}

		return multisig.PubKeyMultisigThreshold{
			K:       uint(k.Multisig.K),
			PubKeys: pk,
		}, nil
	default:
		return nil, fmt.Errorf("fromproto: key type %T is not supported", k)
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
		return kp, fmt.Errorf("toproto: key type %T is not supported", k)
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
		return nil, fmt.Errorf("fromproto: key type %T is not supported", k)
	}
}
