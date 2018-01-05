package auth

import (
	crypto "github.com/tendermint/go-crypto"
)

type SetPubKeyer interface {
	SetPubKey(crypto.PubKey)
}
