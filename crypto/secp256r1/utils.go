package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"io"
	"math/big"
)

func NewPrivateKey(rand io.Reader) {
	_, _ = ecdsa.GenerateKey(pubKeyCurve, rand) // this generates a public & private key pair

}

func PrivKeyFromBytes(curve elliptic.Curve, pk []byte) (*ecdsa.PrivateKey, ecdsa.PublicKey) {
	x, y := curve.ScalarBaseMult(pk)

	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(pk),
	}

	return priv, priv.PublicKey
}
