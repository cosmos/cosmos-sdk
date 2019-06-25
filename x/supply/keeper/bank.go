package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// SendCoinsFromModuleToAccount trasfers coins from a ModuleAccount to an AccAddress
func (k Keeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string,
	recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromModuleToAccount trasfers coins from a ModuleAccount to an AccAddress
func (k Keeper) UndelegateCoinsFromModuleToAccount(ctx sdk.Context, senderModule string,
	recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	_, err := k.bk.UndelegateCoins(ctx, senderAddr, recipientAddr, amt)
	return err
}

// SendCoinsFromModuleToModule trasfers coins from a ModuleAccount to another
func (k Keeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error {

	senderAddr := k.GetModuleAddress(senderModule)
	if senderAddr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if recipientAddr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", recipientModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromAccountToModule trasfers coins from an AccAddress to a ModuleAccount
func (k Keeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress,
	recipientModule string, amt sdk.Coins) sdk.Error {

	recipientAddr := k.GetModuleAddress(recipientModule)
	if recipientAddr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", recipientModule))
	}

	return k.bk.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromAccountToModule trasfers coins from an AccAddress to a ModuleAccount
func (k Keeper) DelegateCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress,
	recipientModule string, amt sdk.Coins) sdk.Error {

	recipientAddr := k.GetModuleAddress(recipientModule)
	if recipientAddr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", recipientModule))
	}

	_, err := k.bk.DelegateCoins(ctx, senderAddr, recipientAddr, amt)
	return err
}

// MintCoins creates new coins from thin air and adds it to the MinterAccount.
// Panics if the name maps to a HolderAccount
func (k Keeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error {
	logger := k.Logger(ctx)

	addr, perm := k.GetModuleAddressAndPermission(moduleName)
	if addr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", moduleName))
	}

	if perm != types.Minter {
		panic(fmt.Sprintf("Account %s does not have permissions to mint tokens", moduleName))
	}

	_, err := k.bk.AddCoins(ctx, addr, amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Inflate(amt)
	k.SetSupply(ctx, supply)

	logger.Info(fmt.Sprintf("minted %s from %s module account", amt, moduleName))

	return nil
}

// BurnCoins burns coins deletes coins from the balance of the module account
func (k Keeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) sdk.Error {
	logger := k.Logger(ctx)

	addr, perm := k.GetModuleAddressAndPermission(moduleName)
	if addr == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", moduleName))
	}

	if perm != types.Burner {
		panic(fmt.Sprintf("Account %s does not have permissions to burn tokens", moduleName))
	}

	_, err := k.bk.SubtractCoins(ctx, addr, amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	logger.Info(fmt.Sprintf("burned %s from %s module account", amt, moduleName))

	return nil
}
