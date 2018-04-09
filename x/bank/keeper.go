package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const moduleName = "bank"

func getCoins(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address) sdk.Coins {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.Coins{}
	}
	return acc.GetCoins()
}

func setCoins(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, amt sdk.Coins) sdk.Error {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		acc = am.NewAccountWithAddress(ctx, addr)
	}
	acc.SetCoins(amt)
	am.SetAccount(ctx, acc)
	return nil
}

// HasCoins returns whether or not an account has at least amt coins.
func hasCoins(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, amt sdk.Coins) bool {
	return getCoins(ctx, am, addr).IsGTE(amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func subtractCoins(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	return newCoins, err
}

// AddCoins adds amt to the coins at the addr.
func addCoins(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Plus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	return newCoins, err
}

// SendCoins moves coins from one account to another
func sendCoins(ctx sdk.Context, am sdk.AccountMapper, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error {
	_, err := subtractCoins(ctx, am, fromAddr, amt)
	if err != nil {
		return err
	}

	_, err = addCoins(ctx, am, toAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// InputOutputCoins handles a list of inputs and outputs
func inputOutputCoins(ctx sdk.Context, am sdk.AccountMapper, inputs []Input, outputs []Output) sdk.Error {
	for _, in := range inputs {
		_, err := subtractCoins(ctx, am, in.Address, in.Coins)
		if err != nil {
			return err
		}
	}

	for _, out := range outputs {
		_, err := addCoins(ctx, am, out.Address, out.Coins)
		if err != nil {
			return err
		}
	}

	return nil
}

// CoinKeeper manages transfers between accounts
type CoinKeeper struct {
	am sdk.AccountMapper
}

// NewCoinKeeper returns a new CoinKeeper
func NewCoinKeeper(am sdk.AccountMapper) CoinKeeper {
	return CoinKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper CoinKeeper) GetCoins(ctx sdk.Context, addr sdk.Address) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// SetCoins sets the coins at the addr.
func (keeper CoinKeeper) SetCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) sdk.Error {
	return setCoins(ctx, keeper.am, addr, amt)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper CoinKeeper) HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func (keeper CoinKeeper) SubtractCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	return subtractCoins(ctx, keeper.am, addr, amt)
}

// AddCoins adds amt to the coins at the addr.
func (keeper CoinKeeper) AddCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) (sdk.Coins, sdk.Error) {
	return addCoins(ctx, keeper.am, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper CoinKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error {
	return sendCoins(ctx, keeper.am, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper CoinKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) sdk.Error {
	return inputOutputCoins(ctx, keeper.am, inputs, outputs)
}

// --------------------------------------------------

// SendKeeper only allows transfers between accounts, without the possibility of creating coins
type SendKeeper struct {
	am sdk.AccountMapper
}

// NewSendKeeper returns a new CoinKeeper
func NewSendKeeper(am sdk.AccountMapper) SendKeeper {
	return SendKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper SendKeeper) GetCoins(ctx sdk.Context, addr sdk.Address) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper SendKeeper) HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper SendKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error {
	return sendCoins(ctx, keeper.am, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper SendKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) sdk.Error {
	return inputOutputCoins(ctx, keeper.am, inputs, outputs)
}

// --------------------------------------------------

// ViewKeeper only allows reading of balances
type ViewKeeper struct {
	am sdk.AccountMapper
}

// NewViewKeeper returns a new CoinKeeper
func NewViewKeeper(am sdk.AccountMapper) ViewKeeper {
	return ViewKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper ViewKeeper) GetCoins(ctx sdk.Context, addr sdk.Address) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper ViewKeeper) HasCoins(ctx sdk.Context, addr sdk.Address, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}
