package crypto

import (
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// public key sentinel errors
var (
	ErrInvalidPubKeySecp256k1Length = errors.New("invalid PubKeySecp256k1 length")
	ErrInvalidPubKeySecp256k1       = errors.New("incompatible PubKeySecp256k1")
	ErrInvalidPubKeyEd25519Length   = errors.New("invalid PubKeyEd25519 length")
	ErrInvalidPubKeyEd25519         = errors.New("incompatible PubKeyEd25519")
)

// GetPubKeySecp256k1 returns a Tendermint secp256k1 public key from the
// PublicKey message type. It will return an error if the size of the public key
// is invalid or the underlying Pub field is invalid.
//
// NOTE: Do not use or call bytes on the result when serializing.
func (m *PublicKey) GetPubKeySecp256k1() (secp256k1.PubKeySecp256k1, error) {
	pk := secp256k1.PubKeySecp256k1{}

	if x, ok := m.GetPub().(*PublicKey_Secp256K1); ok {
		if len(x.Secp256K1) != secp256k1.PubKeySecp256k1Size {
			return pk, ErrInvalidPubKeySecp256k1Length
		}

		copy(pk[:], x.Secp256K1)
		return pk, nil
	}

	return pk, ErrInvalidPubKeySecp256k1
}

// GetPubKeyEd25519 returns a Tendermint Ed25519 public key from the PublicKey
// message type. It will return an error if the size of the public key
// is invalid or the underlying Pub field is invalid.
//
// NOTE: Do not use or call bytes on the result when serializing.
func (m *PublicKey) GetPubKeyEd25519() (ed25519.PubKeyEd25519, error) {
	pk := ed25519.PubKeyEd25519{}

	if x, ok := m.GetPub().(*PublicKey_Ed25519); ok {
		if len(x.Ed25519) != ed25519.PubKeyEd25519Size {
			return pk, ErrInvalidPubKeyEd25519Length
		}

		copy(pk[:], x.Ed25519)
		return pk, nil
	}

	return pk, ErrInvalidPubKeyEd25519
}
