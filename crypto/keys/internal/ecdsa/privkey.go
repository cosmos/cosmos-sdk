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
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{*key}, nil
}

type PrivKey struct {
	ecdsa.PrivateKey
}

// PubKey returns ECDSA public key associated with this private key.
func (sk *PrivKey) PubKey() PubKey {
	return PubKey{sk.PublicKey, nil}
}

// Bytes serialize the private key using big-endian.
func (sk *PrivKey) Bytes() []byte {
	if sk == nil {
		return nil
	}
	fieldSize := (sk.Curve.Params().BitSize + 7) / 8
	bz := make([]byte, fieldSize)
	sk.D.FillBytes(bz)
	return bz
}

// Sign hashes and signs the message usign ECDSA. Implements SDK PrivKey interface.
func (sk *PrivKey) Sign(msg []byte) ([]byte, error) {
	digest := sha256.Sum256(msg)
	return sk.PrivateKey.Sign(rand.Reader, digest[:], nil)
}

// String returns a string representation of the public key based on the curveName.
func (sk *PrivKey) String(name string) string {
	return name + "{-}"
}

// MarshalTo implements proto.Marshaler interface.
func (sk *PrivKey) MarshalTo(dAtA []byte) (int, error) {
	bz := sk.Bytes()
	copy(dAtA, bz)
	return len(bz), nil
}

// Unmarshal implements proto.Marshaler interface.
func (sk *PrivKey) Unmarshal(bz []byte, curve elliptic.Curve, expectedSize int) error {
	if len(bz) != expectedSize {
		return fmt.Errorf("wrong ECDSA SK bytes, expecting %d bytes", expectedSize)
	}

	sk.Curve = curve
	sk.D = new(big.Int).SetBytes(bz)
	sk.X, sk.Y = curve.ScalarBaseMult(bz)
	return nil
}
