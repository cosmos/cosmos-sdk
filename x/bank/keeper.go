package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const moduleName = "bank"

func(am sdk.AccountMapper) AccountMapper {

}


// CoinKeeper manages transfers between accounts
type Keeper struct {
	am sdk.AccountMapper
}

// NewCoinKeeper returns a new CoinKeeper
func NewCoinKeeper(am sdk.AccountMapper) CoinKeeper {
	return CoinKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (ck CoinKeeper) GetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) sdk.Coins {
	acc := ck.am.GetAccount(ctx, addr)
	return acc.GetCoins()
}

// GetCoins returns the coins at the addr.
func (ck CoinKeeper) SetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) sdk.Coins {
	acc := ck.am.GetAccount(ctx, addr)
	acc.SetCoins(amt)
}

// HasCoins returns whether or not an account has at least amt coins.
func (ck CoinKeeper) HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) bool {
	acc := ck.am.GetAccount(ctx, addr)
	return acc.GetCoins().IsGTE(amt)
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

// InputOutputCoins handles a list of inputs and outputs
func (ck CoinKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) sdk.Error {
	for _, in := range inputs {
		_, err := ck.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return err
		}
	}

	for _, out := range outputs {
		_, err := ck.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err
		}
	}

	return nil
}

// --------------------------------------------------

// SendKeeper only allows transfers between accounts, without the possibility of creating coins
type SendKeeper struct {
	am CoinKeeper
}

// SendCoins moves coins from one account to another
func (sk SendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error {
	return sk.ck.SendCoins(ctx, fromAddr, toAddr, amt)
}

// GetCoins returns the coins at the addr.
func (sk SendKeeper) GetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) sdk.Coins {
	return sk.ck.GetCoins(ctx, addr, amt)
}

// HasCoins returns whether or not an account has at least amt coins.
func (sk SendKeeper) HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) bool {
	return sk.ck.HasCoins(ctx, addr, amt)
}

// --------------------------------------------------

// SendKeeper only allows transfers between accounts, without the possibility of creating coins
type ViewKeeper struct {
	ck CoinKeeper
}

// SendCoins moves coins from one account to another
func (sk SendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error {
	return sk.ck.SendCoins(ctx, fromAddr, toAddr, amt)
}

// GetCoins returns the coins at the addr.
func (sk SendKeeper) GetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) sdk.Coins {
	return sk.ck.GetCoins(ctx, addr, amt)
}

// HasCoins returns whether or not an account has at least amt coins.
func (sk SendKeeper) HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) bool {
	return sk.ck.HasCoins(ctx, addr, amt)
}
