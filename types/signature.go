package types

import crypto "github.com/tendermint/go-crypto"

type StdSignature struct {
	crypto.PubKey // optional
	crypto.Signature
	Sequence int64
}
