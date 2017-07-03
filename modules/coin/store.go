package coin

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
)

type Accountant struct {
	Prefix []byte
}

func (a Accountant) GetAccount(store types.KVStore, addr basecoin.Actor) (Account, error) {
	acct, err := loadAccount(store, a.makeKey(addr))
	// for empty accounts, don't return an error, but rather an empty account
	if IsNoAccountErr(err) {
		err = nil
	}
	return acct, err
}

// CheckCoins makes sure there are funds, but doesn't change anything
func (a Accountant) CheckCoins(store types.KVStore, addr basecoin.Actor, coins types.Coins, seq int) (types.Coins, error) {
	acct, err := a.updateCoins(store, addr, coins, seq)
	return acct.Coins, err
}

// ChangeCoins changes the money, returns error if it would be negative
func (a Accountant) ChangeCoins(store types.KVStore, addr basecoin.Actor, coins types.Coins, seq int) (types.Coins, error) {
	acct, err := a.updateCoins(store, addr, coins, seq)
	if err != nil {
		return acct.Coins, err
	}

	err = storeAccount(store, a.makeKey(addr), acct)
	return acct.Coins, err
}

// updateCoins will load the account, make all checks, and return the updated account.
//
// it doesn't save anything, that is up to you to decide (Check/Change Coins)
func (a Accountant) updateCoins(store types.KVStore, addr basecoin.Actor, coins types.Coins, seq int) (acct Account, err error) {
	acct, err = loadAccount(store, a.makeKey(addr))
	if err != nil {
		return acct, err
	}

	// check sequence
	if seq != acct.Sequence+1 {
		return acct, ErrInvalidSequence()
	}
	acct.Sequence += 1

	// check amount
	final := acct.Coins.Minus(coins)
	if !final.IsNonnegative() {
		return acct, ErrInsufficientFunds()
	}

	acct.Coins = final
	return acct, nil
}

func (a Accountant) makeKey(addr basecoin.Actor) []byte {
	key := addr.Bytes()
	if len(a.Prefix) > 0 {
		key = append(a.Prefix, key...)
	}
	return key
}

type Account struct {
	Coins    types.Coins `json:"coins"`
	Sequence int         `json:"seq"`
}

func loadAccount(store types.KVStore, key []byte) (acct Account, err error) {
	data := store.Get(key)
	if len(data) == 0 {
		return acct, ErrNoAccount()
	}
	err = wire.ReadBinaryBytes(data, &acct)
	if err != nil {
		msg := fmt.Sprintf("Error reading account %X", key)
		return acct, errors.ErrInternal(msg)
	}
	return acct, nil
}

func storeAccount(store types.KVStore, key []byte, acct Account) error {
	bin := wire.BinaryBytes(acct)
	store.Set(key, bin)
	return nil // real stores can return error...
}
