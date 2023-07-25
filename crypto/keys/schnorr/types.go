package schnorr

import (
	"go.dedis.ch/kyber/v4"
)

type Suite interface {
	kyber.Group
	kyber.Random
}

type PrivKey struct {
	Key kyber.Scalar
	Suite
}

type PubKey struct {
	Key kyber.Point
	Suite
}
