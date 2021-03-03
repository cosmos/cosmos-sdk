package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// GenPrivKey generates a new secp256r1 private key. It uses operating system randomness.
func GenPrivKey(curve elliptic.Curve) (PrivKey, error) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	return PrivKey{*key}, err
}

type PrivKey struct {
	ecdsa.PrivateKey
}

// PubKey implements Cosmos-SDK PrivKey interface.
func (sk *PrivKey) PubKey() PubKey {
	return PubKey{sk.PublicKey, nil}
}

// Bytes serialize the private key with first byte being the curve type.
func (sk *PrivKey) Bytes() []byte {
	if sk == nil {
		return nil
	}
	fieldSize := (sk.Curve.Params().BitSize + 7) / 8
	bz := make([]byte, fieldSize)
	sk.D.FillBytes(bz)
	return bz
}

// Sign hashes and signs the message usign ECDSA. Implements sdk.PrivKey interface.
func (sk *PrivKey) Sign(msg []byte) ([]byte, error) {
	digest := sha256.Sum256(msg)
	return sk.PrivateKey.Sign(rand.Reader, digest[:], nil)
}

// String implements proto.Message interface.
func (sk *PrivKey) String(name string) string {
	return name + "{-}"
}

// MarshalTo implements ProtoMarshaler interface.
func (sk *PrivKey) MarshalTo(dAtA []byte) (int, error) {
	bz := sk.Bytes()
	copy(dAtA, bz)
	return len(bz), nil
}

// Unmarshal implements ProtoMarshaler interface.
func (sk *PrivKey) Unmarshal(bz []byte, curve elliptic.Curve, expectedSize int) error {
	if len(bz) != expectedSize {
		return fmt.Errorf("wrong ECDSA SK bytes, expecting %d bytes", expectedSize)
	}

	sk.Curve = curve
	sk.D = new(big.Int).SetBytes(bz)
	sk.X, sk.Y = curve.ScalarBaseMult(bz)
	return nil
}
