package ecdsa

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GenPrivKey generates a new secp256r1 private key. It uses operating system randomness.
func GenPrivKey(c elliptic.Curve) (PrivKey, error) {
	return privFromBuffer(c, rand.Reader)
}

// NewPrivKey creates a private key derived from random bytes
// The bytes size must be at least the size of the Curve field. This function is
// deterministic.
func NewPrivKey(c elliptic.Curve, random []byte) (PrivKey, error) {
	if s := c.Params().BitSize; len(random) < s {
		return PrivKey{}, fmt.Errorf("wrong secret lenght, must be at least %v", s)
	}
	return privFromBuffer(c, bytes.NewBuffer(random))
}

// NewPrivKeyFromSecret creates a private key derived for the secret number
// represented in big-endian. The `secret` must be a curve field element.
// This function is deterministic.
func NewPrivKeyFromSecret(c elliptic.Curve, secret []byte) (PrivKey, error) {
	var d = new(big.Int).SetBytes(secret)
	if d.Cmp(c.Params().N) >= 1 {
		return PrivKey{}, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "secret not in the curve base field")
	}
	sk := new(PrivKey)
	return *sk, sk.Unmarshal(secret, c, (c.Params().Params().BitSize+7)/8)
}

func privFromBuffer(c elliptic.Curve, buf io.Reader) (PrivKey, error) {
	key, err := ecdsa.GenerateKey(c, buf)
	if err != nil {
		return PrivKey{}, err
	}
	return PrivKey{*key}, nil
}

// PrivKey is type which paritally iimplements cryptotypes.PrivKey
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

// Unmarshal implements proto.Marshaler interface. bz must have expectedSize bytes
func (sk *PrivKey) Unmarshal(bz []byte, curve elliptic.Curve, expectedSize int) error {
	if len(bz) != expectedSize {
		return fmt.Errorf("wrong ECDSA SK bytes, expecting %d bytes", expectedSize)
	}

	sk.Curve = curve
	sk.D = new(big.Int).SetBytes(bz)
	sk.X, sk.Y = curve.ScalarBaseMult(bz)
	return nil
}
