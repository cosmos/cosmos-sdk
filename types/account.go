package types

import (
	"fmt"

	"github.com/tendermint/go-crypto"
)

type Account struct {
	PubKey   crypto.PubKey // May be nil, if not known.
	Sequence int
	Balance  int64
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
	SetAccount(acc *Account)
}
