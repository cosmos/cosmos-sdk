package store

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/coin"
	crypto "github.com/tendermint/go-crypto"
)

type Coins = coin.Coins

// Coinser can get and set coins
type Coinser interface {
	GetCoins() Coins
	SetCoins(Coins)
}

// CoinStore manages transfers between accounts
type CoinStore struct {
	AccountStore
}

// SubtractCoins subtracts amt from the coins at the addr.
func (cs CoinStore) SubtractCoins(addr crypto.Address, amt Coins) (Coins, error) {
	acc, err := cs.getCoinserAccount(addr)
	if err != nil {
		return amt, err
	} else if acc == nil {
		return amt, fmt.Errorf("Sending account (%s) does not exist", addr)
	}

	coins := acc.GetCoins()
	newCoins := coins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, ErrInsufficientCoins(fmt.Sprintf("%s < %s", coins, amt))
	}

	acc.SetCoins(newCoins)
	cs.SetAccount(acc.(Account))
	return newCoins, nil
}

// AddCoins adds amt to the coins at the addr.
func (cs CoinStore) AddCoins(addr crypto.Address, amt Coins) (Coins, error) {
	acc, err := cs.getCoinserAccount(addr)
	if err != nil {
		return amt, err
	} else if acc == nil {
		acc = cs.AccountStore.NewAccountWithAddress(addr).(Coinser)
	}

	coins := acc.GetCoins()
	newCoins := coins.Plus(amt)

	acc.SetCoins(newCoins)
	cs.SetAccount(acc.(Account))
	return newCoins, nil
}

// get the account as a Coinser. if the account doesn't exist, return nil.
// if it's not a Coinser, return error.
func (cs CoinStore) getCoinserAccount(addr crypto.Address) (Coinser, error) {
	_acc := cs.GetAccount(addr)
	if _acc == nil {
		return nil, nil
	}

	acc, ok := _acc.(Coinser)
	if !ok {
		return nil, fmt.Errorf("Account %s is not a Coinser", addr)
	}
	return acc, nil
}
