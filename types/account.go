package types

import (
	"errors"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() crypto.Address
	SetAddress(crypto.Address) error // errors if already set.

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
	NewAccountWithAddress(ctx Context, addr crypto.Address) Account
	GetAccount(ctx Context, addr crypto.Address) Account
	SetAccount(ctx Context, acc Account)
}

//-----------------------------------------------------------
// BaseAccount

var _ Account = (*BaseAccount)(nil)

// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type BaseAccount struct {
	Address  crypto.Address `json:"address"`
	Coins    Coins          `json:"coins"`
	PubKey   crypto.PubKey  `json:"public_key"`
	Sequence int64          `json:"sequence"`
}

func NewBaseAccountWithAddress(addr crypto.Address) BaseAccount {
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
func (acc BaseAccount) GetAddress() crypto.Address {
	return acc.Address
}

// Implements sdk.Account.
func (acc *BaseAccount) SetAddress(addr crypto.Address) error {
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
	if acc.PubKey != nil {
		return errors.New("cannot override BaseAccount pubkey")
	}
	acc.PubKey = pubKey
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetCoins() Coins {
	return acc.Coins
}

// Implements sdk.Account.
func (acc *BaseAccount) SetCoins(coins Coins) error {
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
	crypto.RegisterWire(cdc)
}
