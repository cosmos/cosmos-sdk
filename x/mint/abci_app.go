package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker inflates every block and updates inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, k.sk.StakingTokenSupply(ctx))
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.bk.InflateSupply(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateNotBondedTokenSupply(ctx, mintedCoin.Amount) // TODO: verify invariance with bank bond denom supply
}
