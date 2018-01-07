package types

import (
	crypto "github.com/tendermint/go-crypto"
)

// AccountStore indexes accounts by address.
type AccountStore interface {
	NewAccountWithAddress(addr crypto.Address) Account
	GetAccount(addr crypto.Address) Account
	SetAccount(acc Account)
}

// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	Address() crypto.Address

	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetSequence() int64
	SetSequence(int64) error

	Get(key interface{}) (value interface{}, err error)
	Set(key interface{}, value interface{}) error
}
