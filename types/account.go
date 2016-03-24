package types

import (
	"fmt"

	"github.com/tendermint/go-crypto"
)

type Account struct {
	PubKey   crypto.PubKey `json:"pub_key"` // May be nil, if not known.
	Sequence int           `json:"sequence"`
	Balance  int64         `json:"balance"`
}

func (acc *Account) Copy() *Account {
	accCopy := *acc
	return &accCopy
}

func (acc *Account) String() string {
	if acc == nil {
		return "nil-Account"
	}
	return fmt.Sprintf("Account{%v %v %v}",
		acc.PubKey, acc.Sequence, acc.Balance)
}

//----------------------------------------

type PrivAccount struct {
	crypto.PrivKey
	Account
}

//----------------------------------------

type AccountGetter interface {
	GetAccount(addr []byte) *Account
}

type AccountGetterSetter interface {
	GetAccount(addr []byte) *Account
	SetAccount(addr []byte, acc *Account)
}

type AccountCacher interface {
	GetAccount(addr []byte) *Account
	SetAccount(addr []byte, acc *Account)
	Sync()
}
