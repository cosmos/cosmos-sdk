package auth

import (
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/x/coin"
)

//-----------------------------------------------------------
// BaseAccount

// BaseAccount - coin account structure
type BaseAccount struct {
	address  crypto.Address
	coins    coin.Coins
	pubKey   crypto.PubKey
	sequence int64
}

func NewBaseAccountWithAddress(addr crypto.Address) *BaseAccount {
	return &BaseAccount{
		address: addr,
	}
}

// BaseAccountWire is the account structure used for serialization
type BaseAccountWire struct {
	Address  crypto.Address `json:"address"`
	Coins    coin.Coins     `json:"coins"`
	PubKey   crypto.PubKey  `json:"public_key"` // can't conflict with PubKey()
	Sequence int64          `json:"sequence"`
}

func (acc *BaseAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(BaseAccountWire{
		Address:  acc.address,
		Coins:    acc.coins,
		PubKey:   acc.pubKey,
		Sequence: acc.sequence,
	})
}

func (acc *BaseAccount) UnmarshalJSON(bz []byte) error {
	accWire := new(BaseAccountWire)
	err := json.Unmarshal(bz, accWire)
	if err != nil {
		return err
	}
	acc.address = accWire.Address
	acc.coins = accWire.Coins
	acc.pubKey = accWire.PubKey
	acc.sequence = accWire.Sequence
	return nil
}

// Implements Account
func (acc *BaseAccount) Get(key interface{}) (value interface{}, err error) {
	switch key.(type) {
	case string:
	}
	return nil, nil
}

// Implements Account
func (acc *BaseAccount) Set(key interface{}, value interface{}) error {
	switch key.(type) {
	case string:
	}
	return nil
}

// Implements Account
func (acc *BaseAccount) Address() crypto.Address {
	// TODO: assert address == pubKey.Address()
	return acc.address
}

// Implements Account
func (acc *BaseAccount) GetPubKey() crypto.PubKey {
	return acc.pubKey
}

func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	acc.pubKey = pubKey
	return nil
}

// Implements coinstore.Coinser
func (acc *BaseAccount) GetCoins() coin.Coins {
	return acc.coins
}

// Implements coinstore.Coinser
func (acc *BaseAccount) SetCoins(coins coin.Coins) error {
	acc.coins = coins
	return nil
}

func (acc *BaseAccount) GetSequence() int64 {
	return acc.sequence
}

func (acc *BaseAccount) SetSequence(seq int64) error {
	acc.sequence = seq
	return nil
}
