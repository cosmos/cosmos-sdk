package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Called every block, process inflation on the first block of every hour
func BeginBlocker(ctx sdk.Context, k Keeper) {

	blockTime := ctx.BlockHeader().Time
	minter := k.GetMinter(ctx)
	if blockTime.Sub(minter.InflationLastTime) < time.Hour { // only mint on the hour!
		return
	}

	params := k.GetParams(ctx)
	totalSupply := k.sk.TotalPower(ctx)
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.InflationLastTime = blockTime
	minter, mintedCoin := minter.ProcessProvisions(params, totalSupply, bondedRatio)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	pool := k.sk.GetPool(ctx)
	pool.LooseTokens = pool.LooseTokens.Add(sdk.NewDecFromInt(mintedCoin.Amount))
	k.sk.SetPool(ctx, pool)
	k.SetMinter(ctx, minter)
}
