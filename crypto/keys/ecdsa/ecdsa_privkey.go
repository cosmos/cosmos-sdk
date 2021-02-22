package ecdsa

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// "github.com/cosmos/cosmos-sdk/codec"
// "github.com/cosmos/cosmos-sdk/types/errors"

type ecdsaSK struct {
	ecdsa.PrivateKey
}

var _ cryptotypes.PrivKey = ecdsaSK{}

// GenSecp256r1 generates a new secp256r1 private key. It uses OS randomness.
func GenSecp256r1() (cryptotypes.PrivKey, error) {
	key, err := ecdsa.GenerateKey(secp256r1, rand.Reader)
	return ecdsaSK{*key}, err
}

// PubKey implements SDK PrivKey interface
func (sk ecdsaSK) PubKey() cryptotypes.PubKey {
	return &ecdsaPK{sk.PublicKey, nil}
}

// Bytes serialize the private key with first byte being the curve type
func (sk ecdsaSK) Bytes() []byte {
	bz := make([]byte, PrivKeySize)
	bz[0] = curveTypes[sk.Curve]
	sk.D.FillBytes(bz[1:])
	return bz
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the keys.
func (sk ecdsaSK) Equals(other cryptotypes.LedgerPrivKey) bool {
	sk2, ok := other.(ecdsaSK)
	if !ok {
		return false
	}
	// return EcEqual(&sk.PrivateKey, &sk2.PrivateKey)

	return sk.PrivateKey.Equal(&sk2.PrivateKey)
}

// TODO: remove
// See PublicKey.Equal for details on how Curve is compared.
func EcEqual(priv *ecdsa.PrivateKey, x crypto.PrivateKey) bool {
	xx, ok := x.(*ecdsa.PrivateKey)
	if !ok {
		fmt.Println("not *ecdsa.PrivateKey")
		return false
	}
	return priv.PublicKey.Equal(&xx.PublicKey) && priv.D.Cmp(xx.D) == 0
}

// Type returns key type name. Implements sdk.PrivKey interface
func (sk ecdsaSK) Type() string {
	return curveNames[sk.Curve]
}

// Sign hashes and signs the message usign ECDSA. Implements sdk.PrivKey interface
func (sk ecdsaSK) Sign(msg []byte) ([]byte, error) {
	digest := sha256.Sum256(msg)
	return sk.PrivateKey.Sign(rand.Reader, digest[:], nil)
}

// **** proto.Message ****

func (ecdsaSK) Reset()        {}
func (ecdsaSK) ProtoMessage() {}
func (sk ecdsaSK) String() string {
	return curveNames[sk.Curve] + "{-}"
}
