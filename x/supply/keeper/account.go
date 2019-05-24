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

// GetPoolAccountByName returns a PoolAccount based on the name
func (k Keeper) GetPoolAccountByName(ctx sdk.Context, name string) types.PoolAccount {
	acc := k.GetAccountByName(ctx, name)
	if acc == nil {
		return nil
	}

	pacc, isPoolAccount := acc.(types.PoolAccount)
	if !isPoolAccount {
		return nil
	}

	return pacc
}

// SetPoolAccount sets the pool account to the auth account store
func (k Keeper) SetPoolAccount(ctx sdk.Context, pacc types.PoolAccount) {
	k.ak.SetAccount(ctx, pacc)
}

// GetCoinsByName returns a PoolAccount's coins
func (k Keeper) GetCoinsByName(ctx sdk.Context, name string) sdk.Coins {
	pacc := k.GetPoolAccountByName(ctx, name)
	if pacc == nil {
		return sdk.Coins(nil)
	}
	return pacc.GetCoins()
}
