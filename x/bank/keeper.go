package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const moduleName = "bank"

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

// SendCoins moves coins from one account to another
func (ck CoinKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error {
	_, err := ck.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	_, err = ck.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	return nil
}
