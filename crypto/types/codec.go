package types

import "github.com/tendermint/tendermint/crypto"

type PublicKeyCodec interface {
	Decode(key *PublicKey) (crypto.PubKey, error)
	Encode(key crypto.PubKey) (*PublicKey, error)
}
