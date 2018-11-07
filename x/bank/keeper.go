package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	costGetCoins      sdk.Gas = 10
	costHasCoins      sdk.Gas = 10
	costSetCoins      sdk.Gas = 100
	costSubtractCoins sdk.Gas = 10
	costAddCoins      sdk.Gas = 10
)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	SendKeeper
	SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}

var _ Keeper = (*BaseKeeper)(nil)

// BaseKeeper manages transfers between accounts. It implements the Keeper
// interface.
type BaseKeeper struct {
	am auth.AccountKeeper
}

// NewBaseKeeper returns a new BaseKeeper
func NewBaseKeeper(am auth.AccountKeeper) BaseKeeper {
	return BaseKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper BaseKeeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// SetCoins sets the coins at the addr.
func (keeper BaseKeeper) SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	return setCoins(ctx, keeper.am, addr, amt)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper BaseKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func (keeper BaseKeeper) SubtractCoins(
	ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins,
) (sdk.Coins, sdk.Tags, sdk.Error) {

	return subtractCoins(ctx, keeper.am, addr, amt)
}

// AddCoins adds amt to the coins at the addr.
func (keeper BaseKeeper) AddCoins(
	ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins,
) (sdk.Coins, sdk.Tags, sdk.Error) {

	return addCoins(ctx, keeper.am, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper BaseKeeper) SendCoins(
	ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins,
) (sdk.Tags, sdk.Error) {

	return sendCoins(ctx, keeper.am, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseKeeper) InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error) {
	return inputOutputCoins(ctx, keeper.am, inputs, outputs)
}

//______________________________________________________________________________________________

// SendKeeper defines a module interface that facilitates the transfer of coins
// between accounts without the possibility of creating coins.
type SendKeeper interface {
	ViewKeeper
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error)
}

var _ SendKeeper = (*BaseSendKeeper)(nil)

// SendKeeper only allows transfers between accounts without the possibility of
// creating coins. It implements the SendKeeper interface.
type BaseSendKeeper struct {
	am auth.AccountKeeper
}

// NewBaseSendKeeper returns a new BaseSendKeeper.
func NewBaseSendKeeper(am auth.AccountKeeper) BaseSendKeeper {
	return BaseSendKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper BaseSendKeeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper BaseSendKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

// SendCoins moves coins from one account to another
func (keeper BaseSendKeeper) SendCoins(
	ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins,
) (sdk.Tags, sdk.Error) {

	return sendCoins(ctx, keeper.am, fromAddr, toAddr, amt)
}

// InputOutputCoins handles a list of inputs and outputs
func (keeper BaseSendKeeper) InputOutputCoins(
	ctx sdk.Context, inputs []Input, outputs []Output,
) (sdk.Tags, sdk.Error) {

	return inputOutputCoins(ctx, keeper.am, inputs, outputs)
}

//______________________________________________________________________________________________

// ViewKeeper defines a module interface that facilitates read only access to
// account balances.
type ViewKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool
}

var _ ViewKeeper = (*BaseViewKeeper)(nil)

// BaseViewKeeper implements a read only keeper implementation of ViewKeeper.
type BaseViewKeeper struct {
	am auth.AccountKeeper
}

// NewBaseViewKeeper returns a new BaseViewKeeper.
func NewBaseViewKeeper(am auth.AccountKeeper) BaseViewKeeper {
	return BaseViewKeeper{am: am}
}

// GetCoins returns the coins at the addr.
func (keeper BaseViewKeeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return getCoins(ctx, keeper.am, addr)
}

// HasCoins returns whether or not an account has at least amt coins.
func (keeper BaseViewKeeper) HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool {
	return hasCoins(ctx, keeper.am, addr, amt)
}

//______________________________________________________________________________________________

func getCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress) sdk.Coins {
	ctx.GasMeter().ConsumeGas(costGetCoins, "getCoins")
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.Coins{}
	}
	return acc.GetCoins()
}

func setCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	ctx.GasMeter().ConsumeGas(costSetCoins, "setCoins")
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		acc = am.NewAccountWithAddress(ctx, addr)
	}
	err := acc.SetCoins(amt)
	if err != nil {
		// Handle w/ #870
		panic(err)
	}
	am.SetAccount(ctx, acc)
	return nil
}

// HasCoins returns whether or not an account has at least amt coins.
func hasCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) bool {
	ctx.GasMeter().ConsumeGas(costHasCoins, "hasCoins")
	return getCoins(ctx, am, addr).IsAllGTE(amt)
}

// SubtractCoins subtracts amt from the coins at the addr.
func subtractCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	ctx.GasMeter().ConsumeGas(costSubtractCoins, "subtractCoins")
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Minus(amt)
	if !newCoins.IsNotNegative() {
		return amt, nil, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	tags := sdk.NewTags("sender", []byte(addr.String()))
	return newCoins, tags, err
}

// AddCoins adds amt to the coins at the addr.
func addCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	ctx.GasMeter().ConsumeGas(costAddCoins, "addCoins")
	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Plus(amt)
	if !newCoins.IsNotNegative() {
		return amt, nil, sdk.ErrInsufficientCoins(fmt.Sprintf("%s < %s", oldCoins, amt))
	}
	err := setCoins(ctx, am, addr, newCoins)
	tags := sdk.NewTags("recipient", []byte(addr.String()))
	return newCoins, tags, err
}

// SendCoins moves coins from one account to another
// NOTE: Make sure to revert state changes from tx on error
func sendCoins(ctx sdk.Context, am auth.AccountKeeper, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error) {
	_, subTags, err := subtractCoins(ctx, am, fromAddr, amt)
	if err != nil {
		return nil, err
	}

	_, addTags, err := addCoins(ctx, am, toAddr, amt)
	if err != nil {
		return nil, err
	}

	return subTags.AppendTags(addTags), nil
}

// InputOutputCoins handles a list of inputs and outputs
// NOTE: Make sure to revert state changes from tx on error
func inputOutputCoins(ctx sdk.Context, am auth.AccountKeeper, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error) {
	allTags := sdk.EmptyTags()

	for _, in := range inputs {
		_, tags, err := subtractCoins(ctx, am, in.Address, in.Coins)
		if err != nil {
			return nil, err
		}
		allTags = allTags.AppendTags(tags)
	}

	for _, out := range outputs {
		_, tags, err := addCoins(ctx, am, out.Address, out.Coins)
		if err != nil {
			return nil, err
		}
		allTags = allTags.AppendTags(tags)
	}

	return allTags, nil
}
