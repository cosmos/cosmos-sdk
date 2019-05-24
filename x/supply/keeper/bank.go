package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// SendCoinsModuleToAccount trasfers coins from a ModuleAccount to an AccAddress
func (k Keeper) SendCoinsModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	senderAcc := k.GetModuleAccountByName(ctx, senderModule)
	if senderAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", senderModule))
	}

	err := bank.SendCoins(ctx, k.ak, senderAcc.GetAddress(), recipientAddr, amt)
	if err != nil {
		return err
	}

	return nil
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

	err := bank.SendCoins(ctx, k.ak, senderAcc.GetAddress(), recipientAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// SendCoinsAccountToModule trasfers coins from an AccAddress to a ModuleAccount
func (k Keeper) SendCoinsAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	recipientAcc := k.GetModuleAccountByName(ctx, recipientModule)
	if recipientAcc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", recipientModule))
	}

	err := bank.SendCoins(ctx, k.ak, senderAddr, recipientAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// MintCoins creates new coins from thin air and adds it to the MinterAccount.
// Panics if the name maps to a HolderAccount
func (k Keeper) MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error {
	macc := k.GetModuleAccountByName(ctx, name)
	if macc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", name))
	}

	macc, isMinterAcc := macc.(*types.ModuleMinterAccount)
	if !isMinterAcc {
		panic(fmt.Sprintf("Account holder %s is not allowed to mint coins", name))
	}

	_, err := bank.AddCoins(ctx, k.ak, macc.GetAddress(), amt)
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
func (k Keeper) BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error {
	macc := k.GetModuleAccountByName(ctx, name)
	if macc == nil {
		return sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", name))
	}

	_, err := bank.SubtractCoins(ctx, k.ak, macc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	return nil
}
