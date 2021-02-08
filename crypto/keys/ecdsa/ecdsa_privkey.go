package secp256r1

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// "github.com/cosmos/cosmos-sdk/codec"
//
// "github.com/cosmos/cosmos-sdk/types/errors"

// GenPrivKey generates a new secp256r1 private key. It uses OS randomness.
func GenPrivKey() *PrivKey {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	return PrivKey{key}, err
}

// Bytes returns the byte representation of the Private Key.
// func (privKey *PrivKey) Bytes() []byte {
// 	return privKey.Key
// }

// Sign signs arbitrary data using ECDSA.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	digest := sha256.Sum256(msg)
	return privKey.Key.Sign(rand.Reader, digest, nil)
}

func (pubKey *PubKey) String() string {
	return fmt.Sprintf("secp256r1{%X}", pubKey.Key)
}

// PubKey returns the public key corresponding to privKey.
func (privKey *PrivKey) PubKey(msg []byte) cryptotypes.PubKey {
	pk := privKey.Key.Public()
	return PubKey{pk}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (privKey *PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	if privKey.Type() != other.Type() {
		return false
	}
	pk2, ok := other.(PrivKey)
	privKey.Key.Equal(other)

	return subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}
