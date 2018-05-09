package bank

import (
	"fmt"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// Keeper manages transfers between accounts
type Keeper struct {
	am auth.AccountMapper
}

// NewKeeper returns a new Keeper
func NewKeeper(am auth.AccountMapper) Keeper {
	return Keeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper Keeper) GetCoins(ctx bapp.Context, addr sdk.Address) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// SetCoins sets the coins at the addr.
func (keeper Keeper) SetCoins(ctx bapp.Context, addr sdk.Address, amt sdk.Coins) bapp.Error {
	return setCoins(ctx, keeper.am, addr, amt)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper Keeper) HasCoins(ctx bapp.Context, addr sdk.Address, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func (keeper Keeper) SubtractCoins(ctx bapp.Context, addr sdk.Address, amt sdk.Coins) (sdk.Coins, bapp.Error) {
	return subtractCoins(ctx, keeper.am, addr, amt)
}

// AddCoins adds amt to the coins at the addr.
func (keeper Keeper) AddCoins(ctx bapp.Context, addr sdk.Address, amt sdk.Coins) (sdk.Coins, bapp.Error) {
	return addCoins(ctx, keeper.am, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper Keeper) SendCoins(ctx bapp.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) bapp.Error {
	return sendCoins(ctx, keeper.am, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper Keeper) InputOutputCoins(ctx bapp.Context, inputs []Input, outputs []Output) bapp.Error {
	return inputOutputCoins(ctx, keeper.am, inputs, outputs)
}

//______________________________________________________________________________________________

// SendKeeper only allows transfers between accounts, without the possibility of creating coins
type SendKeeper struct {
	am auth.AccountMapper
}

// NewSendKeeper returns a new Keeper
func NewSendKeeper(am auth.AccountMapper) SendKeeper {
	return SendKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper SendKeeper) GetCoins(ctx bapp.Context, addr sdk.Address) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper SendKeeper) HasCoins(ctx bapp.Context, addr sdk.Address, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper SendKeeper) SendCoins(ctx bapp.Context, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) bapp.Error {
	return sendCoins(ctx, keeper.am, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper SendKeeper) InputOutputCoins(ctx bapp.Context, inputs []Input, outputs []Output) bapp.Error {
	return inputOutputCoins(ctx, keeper.am, inputs, outputs)
}

//______________________________________________________________________________________________

// ViewKeeper only allows reading of balances
type ViewKeeper struct {
	am auth.AccountMapper
}

// NewViewKeeper returns a new Keeper
func NewViewKeeper(am auth.AccountMapper) ViewKeeper {
	return ViewKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper ViewKeeper) GetCoins(ctx bapp.Context, addr sdk.Address) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper ViewKeeper) HasCoins(ctx bapp.Context, addr sdk.Address, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

//______________________________________________________________________________________________

func getCoins(ctx bapp.Context, am auth.AccountMapper, addr sdk.Address) sdk.Coins {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.Coins{}
	}
	return acc.GetCoins()
}

func setCoins(ctx bapp.Context, am auth.AccountMapper, addr sdk.Address, amt sdk.Coins) bapp.Error {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		acc = am.NewAccountWithAddress(ctx, addr)
	}
	acc.SetCoins(amt)
	am.SetAccount(ctx, acc)
	return nil
}

// HasCoins returns whether or not an account has at least amt coins.
func hasCoins(ctx bapp.Context, am auth.AccountMapper, addr sdk.Address, amt sdk.Coins) bool {
	return getCoins(ctx, am, addr).IsGTE(amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func subtractCoins(ctx bapp.Context, am auth.AccountMapper, addr sdk.Address, amt sdk.Coins) (sdk.Coins, bapp.Error) {
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	return newCoins, err
}

// AddCoins adds amt to the coins at the addr.
func addCoins(ctx bapp.Context, am auth.AccountMapper, addr sdk.Address, amt sdk.Coins) (sdk.Coins, bapp.Error) {
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Plus(amt)
	if !newCoins.IsNotNegative() {
		return amt, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	return newCoins, err
}

// SendCoins moves coins from one account to another
// NOTE: Make sure to revert state changes from tx on error
func sendCoins(ctx bapp.Context, am auth.AccountMapper, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) bapp.Error {
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
// NOTE: Make sure to revert state changes from tx on error
func inputOutputCoins(ctx bapp.Context, am auth.AccountMapper, inputs []Input, outputs []Output) bapp.Error {
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
