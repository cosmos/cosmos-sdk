package auth

import (
	"errors"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//-----------------------------------------------------------
// BaseAccount

// BaseAccount - coin account structure
type BaseAccount struct {
	Address  crypto.Address `json:"address"`
	Coins    sdk.Coins      `json:"coins"`
	PubKey   crypto.PubKey  `json:"public_key"`
	Sequence int64          `json:"sequence"`
}

func NewBaseAccountWithAddress(addr crypto.Address) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// Implements Account
func (acc BaseAccount) Get(key interface{}) (value interface{}, err error) {
	panic("not implemented yet")
}

// Implements Account
func (acc *BaseAccount) Set(key interface{}, value interface{}) error {
	panic("not implemented yet")
}

// Implements Account
func (acc BaseAccount) GetAddress() crypto.Address {
	return acc.Address
}

// Implements Account
func (acc *BaseAccount) SetAddress(addr crypto.Address) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// Implements Account
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	return acc.PubKey
}

// Implements Account
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	if acc.PubKey != nil {
		return errors.New("cannot override BaseAccount pubkey")
	}
	acc.PubKey = pubKey
	return nil
}

// Implements Account
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}

// Implements Account
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

// Implements Account
func (acc *BaseAccount) GetSequence() int64 {
	return acc.Sequence
}

// Implements Account
func (acc *BaseAccount) SetSequence(seq int64) error {
	acc.Sequence = seq
	return nil
}
