package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// GetPools returns the bonded and unbonded tokens pool accounts
func (k Keeper) GetBondedPool(ctx sdk.Context) (bondedPool supply.ModuleAccountI) {
	return k.supplyKeeper.GetModuleAccount(ctx, types.BondedTokensName)
}

// GetPools returns the bonded and unbonded tokens pool accounts
func (k Keeper) GetNotBondedPool(ctx sdk.Context) (notBondedPool supply.ModuleAccountI) {
	return k.supplyKeeper.GetModuleAccount(ctx, types.NotBondedTokensName)
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, bondedTokens sdk.Int) {
	bondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), bondedTokens))
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.BondedTokensName, types.NotBondedTokensName, bondedCoins)
	if err != nil {
		panic(err)
	}
}

// notBondedTokensToBonded transfers coins from the not bonded to the bonded pool within staking
func (k Keeper) notBondedTokensToBonded(ctx sdk.Context, notBondedTokens sdk.Int) {
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), notBondedTokens))
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.NotBondedTokensName, types.BondedTokensName, notBondedCoins)
	if err != nil {
		panic(err)
	}
}

// removeBondedTokens removes coins from the bonded pool module account
func (k Keeper) removeBondedTokens(ctx sdk.Context, amt sdk.Int) sdk.Error {
	bondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), amt))
	return k.supplyKeeper.BurnCoins(ctx, types.BondedTokensName, bondedCoins)
}

// removeNotBondedTokens removes coins from the not bonded pool module account
func (k Keeper) removeNotBondedTokens(ctx sdk.Context, amt sdk.Int) sdk.Error {
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), amt))
	return k.supplyKeeper.BurnCoins(ctx, types.NotBondedTokensName, notBondedCoins)
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k Keeper) TotalBondedTokens(ctx sdk.Context) sdk.Int {
	bondedPool := k.GetBondedPool(ctx)
	return bondedPool.GetCoins().AmountOf(k.BondDenom(ctx))
}

// StakingTokenSupply staking tokens from the total supply
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Int {
	return k.supplyKeeper.GetSupply(ctx).Total.AmountOf(k.BondDenom(ctx))
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	bondedPool := k.GetBondedPool(ctx)

	stakeSupply := k.StakingTokenSupply(ctx)
	if stakeSupply.IsPositive() {
		return bondedPool.GetCoins().AmountOf(k.BondDenom(ctx)).ToDec().QuoInt(stakeSupply)
	}
	return sdk.ZeroDec()
}
