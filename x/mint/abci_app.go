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
	supplier := k.bk.GetSupplier(ctx)
	bondingTokenSupply := supplier.TotalAmountOf(k.sk.BondDenom(ctx))
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, bondingTokenSupply)
	k.SetMinter(ctx, minter)

	// mint coins, add to collected fees, update supply
	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.bk.InflateSupply(ctx, sdk.Coins{mintedCoin})
	// k.sk.InflateSupply(ctx) // TODO: increase not bonded tokens
}
