package auth

import crypto "github.com/tendermint/go-crypto"

type Account interface {
	Get(key interface{}) (value interface{})

	Address() []byte
	PubKey() crypto.PubKey
}

type AccountStore interface {
	GetAccount(addr []byte) Account
	SetAccount(acc Account)
}
