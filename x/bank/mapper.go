package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// CoinMapper manages transfers between accounts
type CoinMapper struct {
	am sdk.AccountMapper
}

// SubtractCoins subtracts amt from the coins at the addr.
func (cm CoinMapper) SubtractCoins(ctx sdk.Context, addr crypto.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	acc := cm.am.GetAccount(ctx, addr)
	if acc == nil {
		return amt, sdk.ErrUnrecognizedAddress(addr)
	}

	coins := acc.GetCoins()
	newCoins := coins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, ErrInsufficientCoins(fmt.Sprintf("%s < %s", coins, amt))
	}

	acc.SetCoins(newCoins)
	cm.am.SetAccount(ctx, acc)
	return newCoins, nil
}

// AddCoins adds amt to the coins at the addr.
func (cm CoinMapper) AddCoins(ctx sdk.Context, addr crypto.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	acc := cm.am.GetAccount(ctx, addr)
	if acc == nil {
		acc = cm.am.NewAccountWithAddress(ctx, addr)
	}

	coins := acc.GetCoins()
	newCoins := coins.Plus(amt)

	acc.SetCoins(newCoins)
	cm.am.SetAccount(ctx, acc)
	return newCoins, nil
}
