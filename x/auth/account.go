package auth

import (
	"encoding/json"

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
	return acc.address
}

// Implements Account
func (acc *BaseAccount) SetAddress(addr crypto.Address) error {
	if acc.address != "" {
		return errors.New("cannot override BaseAccount address")
	}
	acc.address = addr
	return nil
}

// Implements Account
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	return acc.pubKey
}

// Implements Account
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	if acc.pubKey != "" {
		return errors.New("cannot override BaseAccount pubkey")
	}
	acc.pubKey = pubKey
	return nil
}

// Implements Account
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.coins
}

// Implements Account
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.coins = coins
	return nil
}

// Implements Account
func (acc *BaseAccount) GetSequence() int64 {
	return acc.sequence
}

// Implements Account
func (acc *BaseAccount) SetSequence(seq int64) error {
	acc.sequence = seq
	return nil
}
