package auth

import (
	"github.com/cosmos/cosmos-sdk/x/store"
	crypto "github.com/tendermint/go-crypto"
)

var _ Auther = (store.Account)(nil)

type Auther interface {
	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetSequence() int64
	SetSequence(int64) error
}
