package auth

import (
	"errors"

	"github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

//-----------------------------------------------------------
// BaseAccount

var _ sdk.Account = (*BaseAccount)(nil)

// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type BaseAccount struct {
	Address  sdk.Address   `json:"address"`
	Coins    sdk.Coins     `json:"coins"`
	PubKey   crypto.PubKey `json:"public_key"`
	Sequence int64         `json:"sequence"`
}

func NewBaseAccountWithAddress(addr sdk.Address) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// Implements sdk.Account.
func (acc BaseAccount) Get(key interface{}) (value interface{}, err error) {
	panic("not implemented yet")
}

// Implements sdk.Account.
func (acc *BaseAccount) Set(key interface{}, value interface{}) error {
	panic("not implemented yet")
}

// Implements sdk.Account.
func (acc BaseAccount) GetAddress() sdk.Address {
	return acc.Address
}

// Implements sdk.Account.
func (acc *BaseAccount) SetAddress(addr sdk.Address) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// Implements sdk.Account.
func (acc BaseAccount) GetPubKey() crypto.PubKey {
	return acc.PubKey
}

// Implements sdk.Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	if !acc.PubKey.Empty() {
		return errors.New("cannot override BaseAccount pubkey")
	}
	acc.PubKey = pubKey
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}

// Implements sdk.Account.
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetSequence() int64 {
	return acc.Sequence
}

// Implements sdk.Account.
func (acc *BaseAccount) SetSequence(seq int64) error {
	acc.Sequence = seq
	return nil
}

//----------------------------------------
// Wire

func RegisterWireBaseAccount(cdc *wire.Codec) {
	// Register crypto.[PubKey,PrivKey,Signature] types.
	wire.RegisterCrypto(cdc)
}
