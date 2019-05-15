package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// SendCoinsPoolToAccount trasfers coins from a PoolAccount to an AccAddress
func (k Keeper) SendCoinsPoolToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	senderAcc, err := k.GetAccountByName(ctx, senderModule)
	if err != nil {
		return err
	}

	err = bank.SendCoins(ctx, k.ak, senderAcc.GetAddress(), recipientAddr, amt)
	if err != nil {
		return err
	}

	return nil
}

// SendCoinsPoolToPool trasfers coins from a PoolAccount to another
func (k Keeper) SendCoinsPoolToPool(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) sdk.Error {
	senderAcc, err := k.GetAccountByName(ctx, senderModule)
	if err != nil {
		return err
	}

	recipientAcc, err := k.GetAccountByName(ctx, recipientModule)
	if err != nil {
		return err
	}

	err = bank.SendCoins(ctx, k.ak, senderAcc.GetAddress(), recipientAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// SendCoinsAccountToPool trasfers coins from an AccAddress to a PoolAccount
func (k Keeper) SendCoinsAccountToPool(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) sdk.Error {
	recipientAcc, err := k.GetAccountByName(ctx, recipientModule)
	if err != nil {
		return err
	}

	err = bank.SendCoins(ctx, k.ak, senderAddr, recipientAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	return nil
}

// MintCoins creates new coins from thin air and adds it to the MinterAccount.
// Panics if the name maps to a HolderAccount
func (k Keeper) MintCoins(ctx sdk.Context, name string, amt sdk.Coins) sdk.Error {
	poolAcc, err := k.GetAccountByName(ctx, name)
	if err != nil {
		return err
	}

	macc, isMinterAcc := poolAcc.(types.PoolMinterAccount)
	if !isMinterAcc {
		panic(fmt.Sprintf("Account holder %s is not allowed to mint coins", name))
	}

	_, err = bank.AddCoins(ctx, k.ak, macc.GetAddress(), amt)
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
	poolAcc, err := k.GetPoolAccountByName(ctx, name)
	if err != nil {
		return err
	}

	_, err = bank.SubtractCoins(ctx, k.ak, poolAcc.GetAddress(), amt)
	if err != nil {
		return err
	}

	// update total supply
	supply := k.GetSupply(ctx)
	supply.Deflate(amt)
	k.SetSupply(ctx, supply)

	return nil
}
