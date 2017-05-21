package types

import (
	"fmt"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

type Account struct {
	PubKey   crypto.PubKey `json:"pub_key"` // May be nil, if not known.
	Sequence int           `json:"sequence"`
	Balance  Coins         `json:"coins"`
}

func (acc *Account) Copy() *Account {
	if acc == nil {
		return nil
	}
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

type AccountSetter interface {
	SetAccount(addr []byte, acc *Account)
}

type AccountGetterSetter interface {
	GetAccount(addr []byte) *Account
	SetAccount(addr []byte, acc *Account)
}

func AccountKey(addr []byte) []byte {
	return append([]byte("base/a/"), addr...)
}

func GetAccount(store KVStore, addr []byte) *Account {
	data := store.Get(AccountKey(addr))
	if len(data) == 0 {
		return nil
	}
	var acc *Account
	err := wire.ReadBinaryBytes(data, &acc)
	if err != nil {
		panic(fmt.Sprintf("Error reading account %X error: %v",
			data, err.Error()))
	}
	return acc
}

func SetAccount(store KVStore, addr []byte, acc *Account) {
	accBytes := wire.BinaryBytes(acc)
	store.Set(AccountKey(addr), accBytes)
}
