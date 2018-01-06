package types

import (
	crypto "github.com/tendermint/go-crypto"
)

type Address = crypto.Address

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
