package schnorr

import (
	"go.dedis.ch/kyber/v4"
)

type Suite interface {
	kyber.Group
	kyber.Random
}

type PrivKey struct {
	Key    []byte
	Scalar kyber.Scalar
	Suite
}

type PubKey struct {
	Key   []byte
	Point kyber.Point
	Suite
}
