package mint

import (
	"time"

	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
)

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) (createdTokens sdk.Coin) {

	// Process provision inflation
	blockTime := ctx.BlockHeader().Time
	if blockTime.Sub(pool.InflationLastTime) >= time.Hour {
		params := k.GetParams(ctx)
		pool.InflationLastTime = blockTime
		pool = pool.ProcessProvisions(params)
		k.SetPool(ctx, pool)
	}

	return
}
