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

	mintedCoin := minter.NextProvision(params, blockTime)
	minter.LastInflation = blockTime

	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, sdk.NewDecFromInt(mintedCoin.Amount))

	// adjust the inflation, hourly-provision rate every hour
	if blockTime.Sub(minter.LastInflationChange) >= time.Hour {
		totalSupply := k.sk.TotalPower(ctx)
		bondedRatio := k.sk.BondedRatio(ctx)
		minter.Inflation = minter.NextInflationRate(params, bondedRatio)
		minter.LastInflationChange = blockTime
		minter.HourlyProvisions = minter.NextHourlyProvisions(params, totalSupply)
	}

	k.SetMinter(ctx, minter)
}
