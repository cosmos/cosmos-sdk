package types

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"
	ed255192 "github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

type PublicKeyCodec interface {
	Decode(key *PublicKey) (crypto.PubKey, error)
	Encode(key crypto.PubKey) (*PublicKey, error)
}

type DefaultPublicKeyCodec struct{}

var _ PublicKeyCodec = DefaultPublicKeyCodec{}

func (DefaultPublicKeyCodec) Decode(key *PublicKey) (crypto.PubKey, error) {
	panic("implement me")
}

func (DefaultPublicKeyCodec) Encode(key crypto.PubKey) (*PublicKey, error) {
	switch key := key.(type) {
	case secp256k1.PubKeySecp256k1:
		return &PublicKey{Sum: &PublicKey_Secp256K1{Secp256K1: key[:]}}, nil
	case ed255192.PubKeyEd25519:
		return &PublicKey{Sum: &PublicKey_Ed25519{Ed25519: key[:]}}, nil
	case sr25519.PubKeySr25519:
		return &PublicKey{Sum: &PublicKey_Sr25519{Sr25519: key[:]}}, nil
	default:
		return nil, fmt.Errorf("can't encode PubKey of type %T", key)
	}
}
