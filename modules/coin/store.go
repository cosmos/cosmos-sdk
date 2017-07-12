package coin

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

// GetAccount - Get account from store and address
func GetAccount(store state.KVStore, addr basecoin.Actor) (Account, error) {
	acct, err := loadAccount(store, addr.Bytes())

	// for empty accounts, don't return an error, but rather an empty account
	if IsNoAccountErr(err) {
		err = nil
	}
	return acct, err
}

// CheckCoins makes sure there are funds, but doesn't change anything
func CheckCoins(store state.KVStore, addr basecoin.Actor, coins Coins) (Coins, error) {
	acct, err := updateCoins(store, addr, coins)
	return acct.Coins, err
}

// ChangeCoins changes the money, returns error if it would be negative
func ChangeCoins(store state.KVStore, addr basecoin.Actor, coins Coins) (Coins, error) {
	acct, err := updateCoins(store, addr, coins)
	if err != nil {
		return acct.Coins, err
	}

	err = storeAccount(store, addr.Bytes(), acct)
	return acct.Coins, err
}

// updateCoins will load the account, make all checks, and return the updated account.
//
// it doesn't save anything, that is up to you to decide (Check/Change Coins)
func updateCoins(store state.KVStore, addr basecoin.Actor, coins Coins) (acct Account, err error) {
	acct, err = loadAccount(store, addr.Bytes())
	// we can increase an empty account...
	if IsNoAccountErr(err) && coins.IsPositive() {
		err = nil
	}
	if err != nil {
		return acct, err
	}

	// check amount
	final := acct.Coins.Plus(coins)
	if !final.IsNonnegative() {
		return acct, ErrInsufficientFunds()
	}

	acct.Coins = final
	return acct, nil
}

// Account - coin account structure
type Account struct {
	Coins Coins `json:"coins"`
}

func loadAccount(store state.KVStore, key []byte) (acct Account, err error) {
	// fmt.Printf("load:  %X\n", key)
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

func storeAccount(store state.KVStore, key []byte, acct Account) error {
	// fmt.Printf("store: %X\n", key)
	bin := wire.BinaryBytes(acct)
	store.Set(key, bin)
	return nil // real stores can return error...
}
