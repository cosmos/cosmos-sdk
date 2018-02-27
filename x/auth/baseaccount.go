package auth

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

//-----------------------------------------------------------
// BaseAccount

var _ sdk.Account = (*BaseAccount)(nil)

// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type BaseAccount struct {
	Address  crypto.Address `json:"address"`
	Coins    sdk.Coins      `json:"coins"`
	PubKey   crypto.PubKey  `json:"public_key"`
	AccNonce int64          `json:"acc_nonce"`
	TxNonce  int64          `json:"tx_nonce"`
}

func NewBaseAccountWithAddress(addr crypto.Address, accNonce int64) BaseAccount {
	return BaseAccount{
		Address:  addr,
		AccNonce: accNonce,
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
func (acc *BaseAccount) GetCoins() sdk.Coins {
	return acc.Coins
}

// Implements sdk.Account.
func (acc *BaseAccount) SetCoins(coins sdk.Coins) error {
	acc.Coins = coins
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetAccNonce() int64 {
	return acc.AccNonce
}

// Implements sdk.Account.
func (acc *BaseAccount) SetAccNonce(nonce int64) error {
	acc.AccNonce = nonce
	return nil
}

// Implements sdk.Account.
func (acc *BaseAccount) GetTxNonce() int64 {
	return acc.TxNonce
}

// Implements sdk.Account.
func (acc *BaseAccount) SetTxNonce(nonce int64) error {
	acc.TxNonce = nonce
	return nil
}

//----------------------------------------
// Wire

func RegisterWireBaseAccount(cdc *wire.Codec) {
	// Register crypto.[PubKey,PrivKey,Signature] types.
	crypto.RegisterWire(cdc)
}
