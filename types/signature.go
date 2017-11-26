package types

import crypto "github.com/tendermint/go-crypto"

type Signature interface {
	CryptoSig() crypto.Signature
	Sequence() int
}

// StdSignature is a simple way to prevent replay attacks.
// There must be better strategies, but this is simplest.
type StdSignature struct {
	crypto.Signature
	Sequence int
}

func (ss StdSignature) CryptoSig() crypto.Signature {
	return ss.Signature
}

func (ss StdSignature) Sequence() int {
	return ss.Sequence
}
