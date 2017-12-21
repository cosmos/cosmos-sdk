package types

import crypto "github.com/tendermint/go-crypto"

type StdSignature struct {
	crypto.Signature
	Sequence int64
}
