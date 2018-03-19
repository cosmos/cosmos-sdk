package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CoinKeeper manages transfers between accounts
type CoinKeeper struct {
	am sdk.AccountMapper
}

// NewCoinKeeper returns a new CoinKeeper
func NewCoinKeeper(am sdk.AccountMapper) CoinKeeper {
	return CoinKeeper{am: am}
}

// SubtractCoins subtracts amt from the coins at the addr.
func (ck CoinKeeper) SubtractCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	acc := ck.am.GetAccount(ctx, addr)
	if acc == nil {
		return amt, sdk.ErrUnknownAddress(addr.String())
	}

	coins := acc.GetCoins()
	newCoins := coins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", coins, amt))
	}

	acc.SetCoins(newCoins)
	ck.am.SetAccount(ctx, acc)
	return newCoins, nil
}

// AddCoins adds amt to the coins at the addr.
func (ck CoinKeeper) AddCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	acc := ck.am.GetAccount(ctx, addr)
	if acc == nil {
		acc = ck.am.NewAccountWithAddress(ctx, addr)
	}

	coins := acc.GetCoins()
	newCoins := coins.Plus(amt)

	acc.SetCoins(newCoins)
	ck.am.SetAccount(ctx, acc)
	return newCoins, nil
}
