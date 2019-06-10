package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SendCoinsModuleToAccount trasfers coins from a ModuleAccount to an AccAddress
func (k Keeper) SendCoinsModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	senderAcc := k.GetModuleAccountByName(ctx, senderModule)
	if senderAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	return k.bk.SendCoins(ctx, senderAcc.GetAddress(), recipientAddr, amt)
}

// SendCoinsModuleToModule trasfers coins from a ModuleAccount to another
func (k Keeper) SendCoinsModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error {
	senderAcc := k.GetModuleAccountByName(ctx, senderModule)
	if senderAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	recipientAcc := k.GetModuleAccountByName(ctx, recipientModule)
	if recipientAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", recipientModule))
	}

	return k.bk.SendCoins(ctx, senderAcc.GetAddress(), recipientAcc.GetAddress(), amt)
}

// SendCoinsAccountToModule trasfers coins from an AccAddress to a ModuleAccount
func (k Keeper) SendCoinsAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	recipientAcc := k.GetModuleAccountByName(ctx, recipientModule)
	if recipientAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", recipientModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// MintCoins creates new coins from thin air and adds it to the MinterAccount.
// Panics if the name maps to a HolderAccount
func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error {
	macc := k.GetModuleAccountByName(ctx, moduleName)
	if macc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", moduleName))
	}

	if !macc.IsMinter() {
		panic(fmt.Sprintf("Account holder %s is not allowed to mint coins", moduleName))
	}

	_, err := k.bk.AddCoins(ctx, macc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Inflate(amt)
	k.SetSupply(ctx, supply)

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error {
	macc := k.GetModuleAccountByName(ctx, moduleName)
	if macc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", moduleName))
	}

	_, err := k.bk.SubtractCoins(ctx, macc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	return nil
}

// DelegateCoins is a wrapper of the bank keeper's DelegateCoins, sending the coins also to the staking bonded ModuleAccount
func (k Keeper) DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, moduleName string, amt sdk.Coins) sdk.Error {
	_, err := k.bk.DelegateCoins(ctx, addr, amt)
	if err != nil {
		return err
	}

	acc := k.GetAccountByName(ctx, moduleName)
	if acc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", moduleName))
	}

	_, err = k.bk.AddCoins(ctx, acc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// UndelegateCoins is a wrapper of the bank keeper's UndelegateCoins, sending the coins also to the staking not bonded ModuleAccount
func (k Keeper) UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, moduleName string, amt sdk.Coins) sdk.Error {
	_, err := k.bk.UndelegateCoins(ctx, addr, amt)
	if err != nil {
		return err
	}

	err = k.BurnCoins(ctx, moduleName, amt)
	if err != nil {
		return err
	}

	return nil
}
