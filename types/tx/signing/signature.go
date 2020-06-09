package signing

import "github.com/tendermint/tendermint/crypto"

type SignatureV2 struct {
	PubKey crypto.PubKey
	Data   SignatureData
}
