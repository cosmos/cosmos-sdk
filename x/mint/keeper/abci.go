package keeper

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker mints new tokens for the previous block.
func (k Keeper) BeginBlocker(ctx context.Context, ic types.InflationCalculationFn) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// recalculate inflation rate
	totalStakingSupply, err := k.StakingTokenSupply(ctx)
	if err != nil {
		return err
	}

	bondedRatio, err := k.BondedRatio(ctx)
	if err != nil {
		return err
	}

	// update minter's inflation and annual provisions
	minter.Inflation = ic(ctx, minter, params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	if err = k.Minter.Set(ctx, minter); err != nil {
		return err
	}

	// calculate minted coins
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	maxSupply := params.MaxSupply
	totalSupply := k.bankKeeper.GetSupply(ctx, params.MintDenom).Amount // fetch total supply from the bank module

	// if maxSupply is not infinite, check against max_supply parameter
	if !maxSupply.IsZero() {
		if totalSupply.Add(mintedCoins.AmountOf(params.MintDenom)).GT(maxSupply) {
			// calculate the difference between maxSupply and totalSupply
			diff := maxSupply.Sub(totalSupply)
			// mint the difference
			diffCoin := sdk.NewCoin(params.MintDenom, diff)
			diffCoins := sdk.NewCoins(diffCoin)

			// mint coins
			if err := k.MintCoins(ctx, diffCoins); err != nil {
				return err
			}
			mintedCoins = diffCoins
		}
	}

	// mint coins if maxSupply is infinite or total staking supply is less than maxSupply
	if maxSupply.IsZero() || totalSupply.Add(mintedCoins.AmountOf(params.MintDenom)).LT(maxSupply) {
		// mint coins
		if err := k.MintCoins(ctx, mintedCoins); err != nil {
			return err
		}
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		return err
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	return k.EventService.EventManager(ctx).EmitKV(
		types.EventTypeMint,
		event.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
		event.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
		event.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
		event.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
	)
}
