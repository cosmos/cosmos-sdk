package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx context.Context, tokens math.Int) error {
	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, tokens))
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.BondedPoolName, types.NotBondedPoolName, coins)
}

// notBondedTokensToBonded transfers coins from the not bonded to the bonded pool within staking
func (k Keeper) notBondedTokensToBonded(ctx context.Context, tokens math.Int) error {
	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, tokens))
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, types.BondedPoolName, coins)
}

// burnBondedTokens burns coins from the bonded pool module account
func (k Keeper) burnBondedTokens(ctx context.Context, amt math.Int) error {
	if !amt.IsPositive() {
		// skip as no coins need to be burned
		return nil
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, amt))

	return k.bankKeeper.BurnCoins(ctx, k.moduleAccountsService.Address(types.BondedPoolName), coins)
}

// burnNotBondedTokens burns coins from the not bonded pool module account
func (k Keeper) burnNotBondedTokens(ctx context.Context, amt math.Int) error {
	if !amt.IsPositive() {
		// skip as no coins need to be burned
		return nil
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, amt))

	return k.bankKeeper.BurnCoins(ctx, k.moduleAccountsService.Address(types.NotBondedPoolName), coins)
}

// TotalBondedTokens total staking tokens supply which is bonded
func (k Keeper) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	bondedPool := k.moduleAccountsService.Address(types.BondedPoolName)
	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}
	return k.bankKeeper.GetBalance(ctx, bondedPool, bondDenom).Amount, nil
}

// StakingTokenSupply staking tokens from the total supply
func (k Keeper) StakingTokenSupply(ctx context.Context) (math.Int, error) {
	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}
	return k.bankKeeper.GetSupply(ctx, bondDenom).Amount, nil
}

// BondedRatio the fraction of the staking tokens which are currently bonded
func (k Keeper) BondedRatio(ctx context.Context) (math.LegacyDec, error) {
	stakeSupply, err := k.StakingTokenSupply(ctx)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	if stakeSupply.IsPositive() {
		totalBonded, err := k.TotalBondedTokens(ctx)
		if err != nil {
			return math.LegacyZeroDec(), err
		}
		return math.LegacyNewDecFromInt(totalBonded).QuoInt(stakeSupply), nil
	}

	return math.LegacyZeroDec(), nil
}
