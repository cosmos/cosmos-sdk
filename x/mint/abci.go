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

	// Given we cannot calculate average blocktime we dont mint on block 1.
	if ctx.BlockHeight() == 1 || minter.LastBlockTimestamp.IsZero() {
		minter.LastBlockTimestamp = ctx.BlockHeader().Time
		k.SetMinter(ctx, minter)

		return
	}

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
// The formula we use is a cumulated moving average.
// averageBlockTime + ((currentBlockTime - averageBlockTime) / blockHeight - 1)
// We remove 1 to blockheight given we dont have an average on block 1.
func getNewAverageBlockTime(ctx sdk.Context, params types.Params, minter types.Minter) time.Duration {
	if !minter.LastBlockTimestamp.IsZero() { // Comes from Genesis
		currentBlockTime := ctx.BlockTime().Sub(minter.LastBlockTimestamp)
		currentAverage := minter.AverageBlockTime.Nanoseconds() +
			((currentBlockTime.Nanoseconds() - minter.AverageBlockTime.Nanoseconds()) /
				(ctx.BlockHeight() - 1))

		return time.Duration(currentAverage)
	}

	return time.Duration(0)
}
