package types

import (
	"encoding/hex"

	crypto "github.com/tendermint/go-crypto"
)

type Address []byte // TODO: cmn.HexBytes

func (a Address) String() string {
	return hex.EncodeToString(a)
}

type Account interface {
	Address() Address

	PubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetCoins() Coins
	SetCoins(Coins) error

	GetSequence() int64
	SetSequence(int64) error

	Get(key interface{}) (value interface{}, err error)
	Set(key interface{}, value interface{}) error
}

type AccountStore interface {
	NewAccountWithAddress(addr Address) Account
	GetAccount(addr Address) Account
	SetAccount(acc Account)
}
