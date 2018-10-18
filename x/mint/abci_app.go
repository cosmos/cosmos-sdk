package mint

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Called every block, process inflation, update validator set
func BeginBlocker(ctx sdk.Context, k Keeper) {
	// Process provision inflation
	blockTime := ctx.BlockHeader().Time
	minter := k.GetMinter(ctx)
	if blockTime.Sub(minter.InflationLastTime) < time.Hour { // nothing to mint!
		return
	}

	params := k.GetParams(ctx)
	totalSupply := k.sk.TotalPower(ctx)
	bondedRatio := k.sk.BondedRatio(ctx)
	minter.InflationLastTime = blockTime
	minter, mintedCoin := minter.ProcessProvisions(params, totalSupply, bondedRatio)
	k.fck.AddCollectedFees(ctx, sdk.Coins{mintedCoin})
	k.SetMinter(ctx, minter)
}
