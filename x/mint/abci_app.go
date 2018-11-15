package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Called every block, process inflation on the first block of every hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	blockTime := ctx.BlockHeader().Time
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)
	totalSupply := k.sk.TotalPower(ctx)
	bondedRatio := k.sk.BondedRatio(ctx)
	minter, mintedCoin := minter.ProcessProvisions(params, blockTime)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.sk.InflateSupply(ctx, sdk.NewDecFromInt(mintedCoin.Amount))

	// adjust the inflation, hourly-provision rate every hour
	if blockTime.Sub(minter.LastInflationChange) >= time.Hour {
		minter.Inflation = minter.NextInflation(params, bondedRatio)
		minter.LastInflationChange = blockTime
		minter.HourlyProvisions = minter.NextHourlyProvisions(params, totalSupply)
	}

	k.SetMinter(ctx, minter)
}
