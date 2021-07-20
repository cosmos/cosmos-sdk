package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

// returns the curve order for the secp256r1 curve
// NOTE: this is specific to the secp256r1/P256 curve,
// and not taken from the domain params for the key itself
// (which would be a more generic approach for all EC)

func p256Order() *big.Int {
	return elliptic.P256().Params().N
}

// returns half the curve order to restrict
// signatures to low s

func p256OrderDiv2() *big.Int {
	return new(big.Int).Div(p256Order(), new(big.Int).SetUint64(2))
}

// A signature is s-normalized if s
// falls in lower half of curve order

func IsSNormalized(sigS *big.Int) bool {
	return sigS.Cmp(p256OrderDiv2()) != 1
}

// normalize the s value if not already in the
// lower half of curve order

func NormalizeS(sigS *big.Int) *big.Int {
	if IsSNormalized(sigS) {
		return sigS
	} else {
		return new(big.Int).Sub(p256Order(), sigS)
	}
}

// GenPrivKey generates a new secp256r1 private key. It uses operating
// system randomness.
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

// Sign hashes and signs the message using ECDSA. Implements SDK
// PrivKey interface.
// NOTE: this now calls the ecdsa Sign function
// (not method!) directly as the s value of the signature is needed to
// low-s normalize the signature value
// It then ASN DER encodes in exactly the same way as the underlying
// go-lang crypto code does (as of 7/20/2021 anyway)

func (sk *PrivKey) Sign(msg []byte) ([]byte, error) {

	digest := sha256.Sum256(msg)
	r, s, err := ecdsa.Sign(rand.Reader, &sk.PrivateKey, digest[:])

	if err != nil {

		return nil, err

	}

	normS := NormalizeS(s)
	var b cryptobyte.Builder

	b.AddASN1(asn1.SEQUENCE, func(b *cryptobyte.Builder) {

		b.AddASN1BigInt(r)
		b.AddASN1BigInt(normS)
	})

	return b.Bytes()
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
