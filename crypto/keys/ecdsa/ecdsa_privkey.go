package ecdsa

import (
	"crypto/ecdsa"
	"crypto/rand"
	//	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// "github.com/cosmos/cosmos-sdk/codec"
// "github.com/cosmos/cosmos-sdk/types/errors"

type ecdsaSK struct {
	*ecdsa.PrivateKey
}

// GenPrivKey generates a new secp256r1 private key. It uses OS randomness.
// TODO: return cryptotypes.PrivKey
func GenSecp256r1() (ecdsaSK, error) {
	key, err := ecdsa.GenerateKey(secp256r1, rand.Reader)
	return ecdsaSK{key}, err
}

// TODO: change return type
func (sk ecdsaSK) PubKey() ecdsaPK {
	return ecdsaPK{&sk.PublicKey, nil}
}

/*
type LedgerPrivKey interface {
	Bytes() []byte
	Sign(msg []byte) ([]byte, error)
	PubKey() PubKey
	Equals(LedgerPrivKey) bool
	Type() string
    }
*/

/*
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

*/
