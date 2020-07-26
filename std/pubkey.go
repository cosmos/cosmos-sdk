package std

import (
	"fmt"

	"github.com/KiraCore/cosmos-sdk/crypto/types"
	"github.com/KiraCore/cosmos-sdk/crypto/types/multisig"

	"github.com/tendermint/tendermint/crypto"
	ed255192 "github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

// DefaultPublicKeyCodec implements the standard PublicKeyCodec for the SDK which
// supports a standard set of public key types
type DefaultPublicKeyCodec struct{}

var _ types.PublicKeyCodec = DefaultPublicKeyCodec{}

// Decode implements the PublicKeyCodec.Decode method
func (cdc DefaultPublicKeyCodec) Decode(key *types.PublicKey) (crypto.PubKey, error) {
	switch key := key.Sum.(type) {
	case *types.PublicKey_Secp256K1:
		n := len(key.Secp256K1)
		if n != secp256k1.PubKeySecp256k1Size {
			return nil, fmt.Errorf("wrong length %d for secp256k1 public key", n)
		}
		var res secp256k1.PubKeySecp256k1
		copy(res[:], key.Secp256K1)
		return res, nil
	case *types.PublicKey_Ed25519:
		n := len(key.Ed25519)
		if n != ed255192.PubKeyEd25519Size {
			return nil, fmt.Errorf("wrong length %d for ed25519 public key", n)
		}
		var res ed255192.PubKeyEd25519
		copy(res[:], key.Ed25519)
		return res, nil
	case *types.PublicKey_Sr25519:
		n := len(key.Sr25519)
		if n != sr25519.PubKeySr25519Size {
			return nil, fmt.Errorf("wrong length %d for sr25519 public key", n)
		}
		var res sr25519.PubKeySr25519
		copy(res[:], key.Sr25519)

		return res, nil
	case *types.PublicKey_Multisig:
		pubKeys := key.Multisig.PubKeys
		resKeys := make([]crypto.PubKey, len(pubKeys))
		for i, k := range pubKeys {
			dk, err := cdc.Decode(k)
			if err != nil {
				return nil, err
			}
			resKeys[i] = dk
		}

		return multisig.NewPubKeyMultisigThreshold(int(key.Multisig.K), resKeys), nil
	default:
		return nil, fmt.Errorf("can't decode PubKey of type %T. Use a custom PublicKeyCodec instead", key)
	}
}

// Encode implements the PublicKeyCodec.Encode method
func (cdc DefaultPublicKeyCodec) Encode(key crypto.PubKey) (*types.PublicKey, error) {
	switch key := key.(type) {
	case secp256k1.PubKeySecp256k1:
		return &types.PublicKey{Sum: &types.PublicKey_Secp256K1{Secp256K1: key[:]}}, nil
	case ed255192.PubKeyEd25519:
		return &types.PublicKey{Sum: &types.PublicKey_Ed25519{Ed25519: key[:]}}, nil
	case sr25519.PubKeySr25519:
		return &types.PublicKey{Sum: &types.PublicKey_Sr25519{Sr25519: key[:]}}, nil
	case multisig.PubKeyMultisigThreshold:
		pubKeys := key.PubKeys
		resKeys := make([]*types.PublicKey, len(pubKeys))
		for i, k := range pubKeys {
			dk, err := cdc.Encode(k)
			if err != nil {
				return nil, err
			}
			resKeys[i] = dk
		}
		return &types.PublicKey{Sum: &types.PublicKey_Multisig{Multisig: &types.PubKeyMultisigThreshold{
			K:       uint32(key.K),
			PubKeys: resKeys,
		}}}, nil
	default:
		return nil, fmt.Errorf("can't encode PubKey of type %T. Use a custom PublicKeyCodec instead", key)
	}
}
