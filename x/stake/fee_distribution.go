package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle fee distribution to the validators and delegators
func (k Keeper) FeeHandler(ctx sdk.Context, collectedFees sdk.Coins) {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)

	// XXX determine
	candidate := NewCandidate(addrs[0], pks[0], Description{})

	// calculate the proposer reward
	precommitPower := k.GetTotalPrecommitVotingPower(ctx)
	toProposer := coinsMulRat(collectedFees, (sdk.NewRat(1, 100).Add(sdk.NewRat(4, 100).Mul(precommitPower).Quo(pool.BondedShares))))
	candidate.ProposerRewardPool = candidate.ProposerRewardPool.Plus(toProposer)

	toReservePool := coinsMulRat(collectedFees, params.ReservePoolFee)
	pool.ReservePool = pool.ReservePool.Plus(toReservePool)

	distributedReward := (collectedFees.Minus(toProposer)).Minus(toReservePool)
	pool.FeePool = pool.FeePool.Plus(distributedReward)
	pool.SumFeesReceived = pool.SumFeesReceived.Plus(distributedReward)
	pool.RecentFee = distributedReward

	k.setPool(ctx, pool)
}

func coinsMulRat(coins sdk.Coins, rat sdk.Rat) sdk.Coins {
	var res sdk.Coins
	for _, coin := range coins {
		coinMulAmt := rat.Mul(sdk.NewRat(coin.Amount)).Evaluate()
		coinMul := sdk.Coins{{coin.Denom, coinMulAmt}}
		res = res.Plus(coinMul)
	}
	return res
}

//____________________________________________________________________________-

// calculate adjustment changes for a candidate at a height
func CalculateAdjustmentChange(candidate Candidate, pool Pool, height int64) (candidate, pool) {

	heightRat := sdk.NewRat(height)
	lastHeightRat := sdk.NewRat(height - 1)
	candidateFeeCount := candidate.VotingPower.Mul(heightRat)
	poolFeeCount := pool.BondedShares.Mul(heightRat)

	// calculate simple and projected pools
	simplePool := candidateFeeCount.Quo(poolFeeCount).Mul(pool.SumFeesReceived)
	calc1 := candidate.PrevPower.Mul(lastHeightRat).Div(pool.PrevPower.Mul(lastHeightRat)).Mul(pool.PrevFeesReceived)
	calc2 := candidate.Power.Div(pool.Power).Mul(pool.RecentFee)
	projectedPool := calc1 + calc2

	AdjustmentChange := simplePool.Sub(projectedPool)
	candidate.Adjustment += AdjustmentChange
	pool.Adjustment += AdjustmentChange
	return candidate, pool
}
