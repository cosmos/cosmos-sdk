package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
	"github.com/tendermint/tendermint/crypto"
)

// GetAccountByName returns an Account based on the name
func (k Keeper) GetAccountByName(ctx sdk.Context, name string) auth.Account {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	return k.ak.GetAccount(ctx, moduleAddress)
}

// GetModuleAccountByName returns a ModuleAccount based on the name
func (k Keeper) GetModuleAccountByName(ctx sdk.Context, name string) types.ModuleAccount {
	acc := k.GetAccountByName(ctx, name)
	if acc == nil {
		return nil
	}

	macc, isModuleAccount := acc.(types.ModuleAccount)
	if !isModuleAccount {
		return nil
	}

	return macc
}

// SetModuleAccount sets the pool account to the auth account store
func (k Keeper) SetModuleAccount(ctx sdk.Context, macc types.ModuleAccount) {
	k.ak.SetAccount(ctx, macc)
}

// GetCoins alias for bank keeper
func (k Keeper) GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return k.bk.GetCoins(ctx, addr)
}

// GetCoinsByName returns a ModuleAccount's coins
func (k Keeper) GetCoinsByName(ctx sdk.Context, name string) sdk.Coins {
	macc := k.GetModuleAccountByName(ctx, name)
	if macc == nil {
		return sdk.Coins(nil)
	}
	return macc.GetCoins()
}
