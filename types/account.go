package types

import (
	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

// Address in go-crypto style
type Address = cmn.HexBytes

// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() Address
	SetAddress(Address) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetSequence() int64
	SetSequence(int64) error

	GetCoins() Coins
	SetCoins(Coins) error

	Get(key interface{}) (value interface{}, err error)
	Set(key interface{}, value interface{}) error
}

// AccountMapper stores and retrieves accounts from stores
// retrieved from the context.
type AccountMapper interface {
	NewAccountWithAddress(ctx Context, addr Address) Account
	GetAccount(ctx Context, addr Address) Account
	SetAccount(ctx Context, acc Account)
}

// AccountDecoder unmarshals account bytes
type AccountDecoder func(accountBytes []byte) (Account, error)
