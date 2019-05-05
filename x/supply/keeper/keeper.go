package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
	"github.com/tendermint/tendermint/crypto"
)

// SendKeeper
type SendKeeper interface {
	bank.ViewKeeper // GetCoins, HasCoins, Codespace

	SendCoinsModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error
	SendCoinsModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error
	SendCoinsAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error
	MintCoins(ctx sdk.Context, module string, amt sdk.Coins) sdk.Error // panics if used with with holder account

	GetSendEnabled(ctx sdk.Context) bool
	SetSendEnabled(ctx sdk.Context, enabled bool)
}

//-----------------------------------------------------------------------------
// BaseSendKeeper

var _ SendKeeper = (*BaseSendKeeper)(nil)

// BaseSendKeeper
type BaseSendKeeper struct {
	*bank.BaseViewKeeper

	ak         auth.AccountKeeper
	paramSpace params.Subspace
}

// NewBaseSendKeeper
func NewBaseSendKeeper(ak auth.AccountKeeper, codespace sdk.CodespaceType, paramSpace params.Subspace) BaseSendKeeper {
	baseViewKeeper := bank.NewBaseViewKeeper(ak, codespace)
	return BaseSendKeeper{
		&baseViewKeeper,
		ak,
		paramSpace,
	}
}

// SendCoinsModuleToAccount
func (hk BaseSendKeeper) SendCoinsModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	senderAcc, err := getModuleAccount(ctx, hk.ak, senderModule)
	if err != nil {
		return err
	}

	err = sendCoins(ctx, hk.ak, senderAcc.GetAddress(), recipientAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// SendCoinsModuleToModule
func (hk BaseSendKeeper) SendCoinsModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error {
	senderAcc, err := getModuleAccount(ctx, hk.ak, senderModule)
	if err != nil {
		return err
	}

	recipientAcc, err := getModuleAccount(ctx, hk.ak, recipientModule)
	if err != nil {
		return err
	}

	err = sendCoins(ctx, hk.ak, senderAcc.GetAddress(), recipientAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// SendCoinsAccountToModule
func (hk BaseSendKeeper) SendCoinsAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	recipientAcc, err := getModuleAccount(ctx, hk.ak, recipientModule)
	if err != nil {
		return err
	}

	err = sendCoins(ctx, hk.ak, senderAddr, recipientAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// MintCoins
func (hk BaseSendKeeper) MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error {
	moduleAcc, err := getModuleAccount(ctx, hk.ak, name)
	if err != nil {
		return err
	}

	macc, isMinterAcc := moduleAcc.(types.ModuleMinterAccount)
	if !isMinterAcc {
		panic(fmt.Sprintf("Account holder %s is not allowed to mint coins", name))
	}

	_, err = addCoins(ctx, hk.ak, macc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// GetSendEnabled
func (hk BaseSendKeeper) GetSendEnabled(ctx sdk.Context) (enabled bool) {
	hk.paramSpace.Get(ctx, ParamStoreKeySendEnabled, &enabled)
	return
}

// SetSendEnabled
func (hk BaseSendKeeper) SetSendEnabled(ctx sdk.Context, enabled bool) {
	hk.paramSpace.Set(ctx, ParamStoreKeySendEnabled, &enabled)
}

//-----------------------------------------------------------------------------
// private functions

func getModuleAccount(ctx sdk.Context, ak auth.AccountKeeper, name string) (auth.Account, sdk.Error) {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	acc := ak.GetAccount(ctx, moduleAddress)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", name))
	}
	return acc, nil
}

//-----------------------------------------------------------------------------
// private functions from bank module

func getCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress) sdk.Coins {
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		return sdk.NewCoins()
	}
	return acc.GetCoins()
}

func setCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	if !amt.IsValid() {
		return sdk.ErrInvalidCoins(amt.String())
	}
	acc := am.GetAccount(ctx, addr)
	if acc == nil {
		acc = am.NewAccountWithAddress(ctx, addr)
	}
	err := acc.SetCoins(amt)
	if err != nil {
		panic(err)
	}
	am.SetAccount(ctx, acc)
	return nil
}

// HasCoins returns whether or not an account has at least amt coins.
func hasCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) bool {
	return getCoins(ctx, am, addr).IsAllGTE(amt)
}

// subtractCoins subtracts amt coins from an account with the given address addr.
//
// CONTRACT: If the account is a vesting account, the amount has to be spendable.
func subtractCoins(ctx sdk.Context, ak auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error) {

	if !amt.IsValid() {
		return nil, sdk.ErrInvalidCoins(amt.String())
	}

	oldCoins, spendableCoins := sdk.NewCoins(), sdk.NewCoins()

	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		oldCoins = acc.GetCoins()
		spendableCoins = acc.SpendableCoins(ctx.BlockHeader().Time)
	}

	// For non-vesting accounts, spendable coins will simply be the original coins.
	// So the check here is sufficient instead of subtracting from oldCoins.
	_, hasNeg := spendableCoins.SafeSub(amt)
	if hasNeg {
		return amt, sdk.ErrInsufficientCoins(
			fmt.Sprintf("insufficient account funds; %s < %s", spendableCoins, amt),
		)
	}

	newCoins := oldCoins.Sub(amt) // should not panic as spendable coins was already checked
	err := setCoins(ctx, ak, addr, newCoins)

	return newCoins, err
}

// AddCoins adds amt to the coins at the addr.
func addCoins(ctx sdk.Context, am auth.AccountKeeper, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Error) {

	if !amt.IsValid() {
		return nil, sdk.ErrInvalidCoins(amt.String())
	}

	oldCoins := getCoins(ctx, am, addr)
	newCoins := oldCoins.Add(amt)

	if newCoins.IsAnyNegative() {
		return amt, sdk.ErrInsufficientCoins(
			fmt.Sprintf("insufficient account funds; %s < %s", oldCoins, amt),
		)
	}

	err := setCoins(ctx, am, addr, newCoins)

	return newCoins, err
}

// SendCoins moves coins from one account to another
// Returns ErrInvalidCoins if amt is invalid.
func sendCoins(ctx sdk.Context, am auth.AccountKeeper, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
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
