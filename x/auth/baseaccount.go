package auth

import (
	"errors"

	"github.com/tendermint/go-crypto"

	bam "github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/cosmos/cosmos-sdk/wire"
)

//-----------------------------------------------------------
// BaseAccount

var _ Account = (*BaseAccount)(nil)

// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type BaseAccount struct {
	Address  bam.Address   `json:"address"`
	Coins    bam.Coins     `json:"coins"`
	PubKey   crypto.PubKey `json:"public_key"`
	Sequence int64         `json:"sequence"`
}

func NewBaseAccountWithAddress(addr bam.Address) BaseAccount {
	return BaseAccount{
		Address: addr,
	}
}

// Implements Account.
func (acc *BaseAccount) GetAddress() bam.Address {
	return acc.Address
}

// Implements Account.
func (acc *BaseAccount) SetAddress(addr bam.Address) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}
	acc.Address = addr
	return nil
}

// Implements Account.
func (acc *BaseAccount) GetPubKey() crypto.PubKey {
	return acc.PubKey
}

// Implements Account.
func (acc *BaseAccount) SetPubKey(pubKey crypto.PubKey) error {
	if acc.PubKey != nil {
		return errors.New("cannot override BaseAccount pubkey")
	}
	acc.PubKey = pubKey
	return nil
}

// Implements Account.
func (acc *BaseAccount) GetCoins() bam.Coins {
	return acc.Coins
}

// Implements Account.
func (acc *BaseAccount) SetCoins(coins bam.Coins) error {
	acc.Coins = coins
	return nil
}

// Implements Account.
func (acc *BaseAccount) GetSequence() int64 {
	return acc.Sequence
}

// Implements Account.
func (acc *BaseAccount) SetSequence(seq int64) error {
	acc.Sequence = seq
	return nil
}

//----------------------------------------
// Wire

// Most users shouldn't use this, but this comes handy for tests.
func RegisterBaseAccount(cdc *wire.Codec) {
	cdc.RegisterInterface((*Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/BaseAccount", nil)
	wire.RegisterCrypto(cdc)
}
