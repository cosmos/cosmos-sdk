package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SendCoinsFromModuleToAccount trasfers coins from a ModuleAccount to an AccAddress
func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	senderAcc := k.GetModuleAccountByName(ctx, senderModule)
	if senderAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	return k.bk.SendCoins(ctx, senderAcc.GetAddress(), recipientAddr, amt)
}

// SendCoinsFromModuleToModule trasfers coins from a ModuleAccount to another
func (k Keeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error {
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

// SendCoinsFromAccountToModule trasfers coins from an AccAddress to a ModuleAccount
func (k Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
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
