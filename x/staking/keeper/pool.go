package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// names used as root for pool module accounts:
//
// - NotBondedPool -> "NotBondedTokens"
//
// - BondedPool -> "BondedTokens"
const (
	NotBondedTokensName = "NotBondedTokensPool"
	BondedTokensName    = "BondedTokensPool"
)

// GetPools returns the bonded and unbonded tokens pool accounts
func (k Keeper) GetBondedPool(ctx sdk.Context) (bondedPool supply.ModuleAccount) {
	bondedPool = k.supplyKeeper.GetModuleAccountByName(ctx, BondedTokensName)
	return bondedPool
}

// GetPools returns the bonded and unbonded tokens pool accounts
func (k Keeper) GetNotBondedPool(ctx sdk.Context) (notBondedPool supply.ModuleAccount) {
	notBondedPool = k.supplyKeeper.GetModuleAccountByName(ctx, NotBondedTokensName)
	return notBondedPool
}

// SetBondedPool sets the bonded tokens pool account
func (k Keeper) SetBondedPool(ctx sdk.Context, bondedPool supply.ModuleAccount) {
	// safety check
	if bondedPool.Name() != BondedTokensName {
		panic(fmt.Sprintf("invalid name for bonded pool (%s ≠ %s)", BondedTokensName, bondedPool.Name()))
	}
	k.supplyKeeper.SetModuleAccount(ctx, bondedPool)
}

// SetNotBondedPool sets the not bonded tokens pool account
func (k Keeper) SetNotBondedPool(ctx sdk.Context, notBondedPool supply.ModuleAccount) {
	// safety check
	if notBondedPool.Name() != NotBondedTokensName {
		panic(fmt.Sprintf("invalid name for unbonded pool (%s ≠ %s)", NotBondedTokensName, notBondedPool.Name()))
	}
	k.supplyKeeper.SetModuleAccount(ctx, notBondedPool)
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, bondedTokens sdk.Int) {
	bondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), bondedTokens))
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, BondedTokensName, NotBondedTokensName, bondedCoins)
	if err != nil {
		panic(err)
	}
}

// notBondedTokensToBonded transfers coins from the not bonded to the bonded pool within staking
func (k Keeper) notBondedTokensToBonded(ctx sdk.Context, notBondedTokens sdk.Int) {
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), notBondedTokens))
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, NotBondedTokensName, BondedTokensName, notBondedCoins)
	if err != nil {
		panic(err)
	}
}

// removeBondedTokens removes coins from the bonded pool module account
func (k Keeper) removeBondedTokens(ctx sdk.Context, amt sdk.Int) sdk.Error {
	bondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), amt))
	return k.supplyKeeper.BurnCoins(ctx, BondedTokensName, bondedCoins)
}

// removeNotBondedTokens removes coins from the not bonded pool module account
func (k Keeper) removeNotBondedTokens(ctx sdk.Context, amt sdk.Int) sdk.Error {
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), amt))
	return k.supplyKeeper.BurnCoins(ctx, NotBondedTokensName, notBondedCoins)
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
