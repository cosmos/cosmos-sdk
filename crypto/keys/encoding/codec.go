package encoding

import (
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	pc "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

// PubKeyToProto takes crypto.PubKey and transforms it to a protobuf Pubkey
func PubKeyToProto(k crypto.PubKey) (pc.PublicKey, error) {
	var kp pc.PublicKey
	switch k := k.(type) {
	case ed25519.PubKey:
		kp = pc.PublicKey{
			Sum: &pc.PublicKey_Ed25519{
				Ed25519: k,
			},
		}
	default:
		return kp, fmt.Errorf("toproto: key type %v is not supported", k)
	}
	return kp, nil
}

// PubKeyFromProto takes a protobuf Pubkey and transforms it to a crypto.Pubkey
func PubKeyFromProto(k pc.PublicKey) (crypto.PubKey, error) {
	switch k := k.Sum.(type) {
	case *pc.PublicKey_Ed25519:
		if len(k.Ed25519) != ed25519.PubKeySize {
			return nil, fmt.Errorf("invalid size for PubKeyEd25519. Got %d, expected %d",
				len(k.Ed25519), ed25519.PubKeySize)
		}
		pk := make(ed25519.PubKey, ed25519.PubKeySize)
		copy(pk, k.Ed25519)
		return pk, nil
	default:
		return nil, fmt.Errorf("fromproto: key type %v is not supported", k)
	}
}

// PrivKeyToProto takes crypto.PrivKey and transforms it to a protobuf PrivKey
func PrivKeyToProto(k crypto.PrivKey) (pc.PrivateKey, error) {
	var kp pc.PrivateKey
	switch k := k.(type) {
	case ed25519.PrivKey:
		kp = pc.PrivateKey{
			Sum: &pc.PrivateKey_Ed25519{
				Ed25519: k,
			},
		}
	default:
		return kp, errors.New("toproto: key type is not supported")
	}
	return kp, nil
}

// PrivKeyFromProto takes a protobuf PrivateKey and transforms it to a crypto.PrivKey
func PrivKeyFromProto(k pc.PrivateKey) (crypto.PrivKey, error) {
	switch k := k.Sum.(type) {
	case *pc.PrivateKey_Ed25519:

		if len(k.Ed25519) != ed25519.PubKeySize {
			return nil, fmt.Errorf("invalid size for PubKeyEd25519. Got %d, expected %d",
				len(k.Ed25519), ed25519.PubKeySize)
		}
		pk := make(ed25519.PrivKey, ed25519.PrivateKeySize)
		copy(pk, k.Ed25519)
		return pk, nil
	default:
		return nil, errors.New("fromproto: key type not supported")
	}
}
