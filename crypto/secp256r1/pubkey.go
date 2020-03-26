package secp256r1

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

var _ crypto.PubKey = PubKeyNistp256{}

// PubKeySr25519Size is the number of bytes in an Sr25519 public key.
const PubKeySr25519Size = 32

// PubKeySr25519 implements crypto.PubKey for the Sr25519 signature scheme.
type PubKeyNistp256 [PubKeySr25519Size]byte

func (pubKey PubKeyNistp256) Address() crypto.Address {
	return crypto.Address(tmhash.SumTruncated(pubKey[:]))
}

func (pubKey PubKeyNistp256) Bytes() []byte {
	bz, err := cdc.MarshalBinaryBare(pubKey)
	if err != nil {
		panic(err)
	}
	return bz
}

func (pubKey PubKeyNistp256) VerifyBytes(msg []byte, sig []byte) bool {

}

func (pubKey PubKeyNistp256) String() string {

}

func (pubKey PubKeyNistp256) Equals(other crypto.PubKey) bool {

}
