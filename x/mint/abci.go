package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	minter.AverageBlockTime = getNewAverageBlockTime(ctx, params, minter)

	// recalculate inflation rate
	totalStakingSupply := k.StakingTokenSupply(ctx)
	bondedRatio := k.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	minter.LastBlockTimestamp = ctx.BlockHeader().Time
	k.SetMinter(ctx, minter)

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}

// getNewAverageBlockTime calculates the average time it takes to forge a new block based on the cumulated
// moving average since the beginning of the chain.
// On first block we use params.BlocksPerYear to calculate it, so the average in block one is:
// NanosecondsInAYear / params.BlocksPErYear
func getNewAverageBlockTime(ctx sdk.Context, params types.Params, minter types.Minter) time.Duration {
	if minter.LastBlockTimestamp.IsZero() { // Comes from Genesis
		// Calculate Average by BlocksPerYear param.
		nanoSecondsABlock := types.YEAR.Nanoseconds() / int64(params.BlocksPerYear)
		return time.Duration(nanoSecondsABlock)
	} else {
		currentBlockTime := ctx.BlockHeader().Time.Sub(minter.LastBlockTimestamp)
		// Calculate Cumulative moving average of block time.
		currentAverage := (currentBlockTime.Nanoseconds() + minter.AverageBlockTime.Nanoseconds()*(ctx.BlockHeight()-1)) / ctx.BlockHeight()
		return time.Duration(currentAverage)
	}
}
