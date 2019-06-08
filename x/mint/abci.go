package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	totalSupply := k.TotalTokens(ctx)
	bondedRatio := k.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.InflateSupply(ctx, mintedCoin.Amount)
}
