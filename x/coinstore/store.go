package coin

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

type Coins = types.Coins

type Coinser interface {
	GetCoins() Coins
	SetCoins(Coins)
}

// CoinStore manages transfers between accounts
type CoinStore struct {
	types.AccountStore
}

// get the account as a Coinser. if the account doesn't exist, return nil.
// if it's not a Coinser, return error.
func (cs CoinStore) getCoinserAccount(addr types.Address) (types.Coinser, error) {
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

func (cs CoinStore) SubtractCoins(addr types.Address, amt Coins) (Coins, error) {
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
	cs.SetAccount(acc.(types.Account))
	return newCoins, nil
}

func (cs CoinStore) AddCoins(addr types.Address, amt Coins) (Coins, error) {
	acc, err := cs.getCoinserAccount(addr)
	if err != nil {
		return amt, err
	} else if acc == nil {
		acc = cs.AccountStore.NewAccountWithAddress(addr).(Coinser)
	}

	coins := acc.GetCoins()
	newCoins := coins.Plus(amt)

	acc.SetCoins(newCoins)
	cs.SetAccount(acc.(types.Account))
	return newCoins, nil
}

/*
// TransferCoins transfers coins from fromAddr to toAddr.
// It returns an error if the from account doesn't exist,
// if the accounts doin't implement Coinser,
// or if the from account does not have enough coins.
func (cs CoinStore) TransferCoins(fromAddr, toAddr types.Address, amt Coins) error {
	var fromAcc, toAcc types.Account

	// Get the accounts
	_fromAcc := cs.GetAccount(fromAddr)
	if _fromAcc == nil {
		return ErrUnknownAccount(fromAddr)
	}

	_toAcc := cs.GetAccount(to)
	if _toAcc == nil {
		toAcc = cs.AccountStore.NewAccountWithAddress(to)
	}

	// Ensure they are Coinser
	fromAcc, ok := _fromAcc.(Coinser)
	if !ok {
		return ErrAccountNotCoinser(from)
	}

	toAcc, ok = _toAcc.(Coinser)
	if !ok {
		return ErrAccountNotCoinser(from)
	}

	// Coin math
	fromCoins := fromAcc.GetCoins()
	newFromCoins := fromCoins.Minus(amt)
	if newFromCoins.Negative() {
		return ErrInsufficientCoins(fromCoins, amt)
	}
	toCoins := toAcc.GetCoins()
	newToCoins := toCoins.Plus(amt)

	// Set everything!
	fromAcc.SetCoins(newFromCoins)
	toAcc.SetCoins(newToCoins)
	cs.SetAccount(fromAcc)
	cs.SetAccount(toAcc)

	return nil
}*/
