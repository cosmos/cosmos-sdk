package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Inflate every block, update inflation parameters once per hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	blockTime := ctx.BlockHeader().Time
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, sdk.NewDecFromInt(mintedCoin.Amount))

	if blockTime.Sub(minter.LastUpdate) < time.Hour {
		return
	}

	// adjust the inflation, hourly-provision rate every hour
	totalSupply := k.sk.TotalPower(ctx)
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	minter.LastUpdate = blockTime
	k.SetMinter(ctx, minter)
}
