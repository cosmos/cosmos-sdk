package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() sdk.Address
	SetAddress(sdk.Address) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetSequence() int64
	SetSequence(int64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error
}

// AccountDecoder unmarshals account bytes
type AccountDecoder func(accountBytes []byte) (Account, error)
