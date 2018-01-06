package main

import (
	"github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// AppAccount - coin account structure
type AppAccount struct {
	Address_ types.Address `json:"address"`
	Coins    types.Coins   `json:"coins"`
	PubKey_  crypto.PubKey `json:"public_key"` // can't conflict with PubKey()
	Sequence int64         `json:"sequence"`
}

// Implements auth.Account
func (a *AppAccount) Get(key interface{}) (value interface{}, err error) {
	switch key.(type) {
	case string:
		//
	default:
		panic("HURAH!")
	}
	return nil, nil
}

// Implements auth.Account
func (a *AppAccount) Set(key interface{}, value interface{}) error {
	switch key.(type) {
	case string:
		//
	default:
		panic("HURAH!")
	}
	return nil
}

// Implements auth.Account
func (a *AppAccount) Address() types.Address {
	return a.PubKey_.Address()
}

// Implements auth.Account
func (a *AppAccount) PubKey() crypto.PubKey {
	return a.PubKey_
}

func (a *AppAccount) SetPubKey(pubKey crypto.PubKey) error {
	a.PubKey_ = pubKey
	return nil
}

// Implements coinstore.Coinser
func (a *AppAccount) GetCoins() types.Coins {
	return a.Coins
}

// Implements coinstore.Coinser
func (a *AppAccount) SetCoins(coins types.Coins) error {
	a.Coins = coins
	return nil
}

func (a *AppAccount) GetSequence() int64 {
	return a.Sequence
}

func (a *AppAccount) SetSequence(seq int64) error {
	a.Sequence = seq
	return nil
}
