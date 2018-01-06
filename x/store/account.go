package store

import (
	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/x/coin"
)

// AccountStore indexes accounts by address.
type AccountStore interface {
	NewAccountWithAddress(addr crypto.Address) Account
	GetAccount(addr crypto.Address) Account
	SetAccount(acc Account)
}

// Account is a standard balance account
// using a sequence number for replay protection
// and a single pubkey for authentication.
// TODO: multisig accounts?
type Account interface {
	Address() crypto.Address

	GetPubKey() crypto.PubKey
	SetPubKey(crypto.PubKey) error

	GetCoins() coin.Coins
	SetCoins(coin.Coins) error

	GetSequence() int64
	SetSequence(int64) error

	Get(key interface{}) (value interface{}, err error)
	Set(key interface{}, value interface{}) error
}
