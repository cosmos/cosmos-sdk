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

// XXX need to introduce rat amount based coins for the pool :(
func coinsMulRat(coins sdk.Coins, rat sdk.Rat) sdk.Coins {
	var res sdk.Coins
	for _, coin := range coins {
		coinMulAmt := rat.Mul(sdk.NewRat(coin.Amount)).Evaluate()
		coinMul := sdk.Coins{{coin.Denom, coinMulAmt}}
		res = res.Plus(coinMul)
	}
	return res
}
