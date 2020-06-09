package types

import (
	"github.com/tendermint/tendermint/crypto"
)

// PublicKeyCodec defines a type which can encode and decode crypto.PubKey's
// to and from protobuf PublicKey's
type PublicKeyCodec interface {
	// Encode encodes the crypto.PubKey as a protobuf PublicKey or returns an error
	Encode(key crypto.PubKey) (*PublicKey, error)

	// Decode decodes a crypto.PubKey from a protobuf PublicKey or returns an error
	Decode(key *PublicKey) (crypto.PubKey, error)
}
