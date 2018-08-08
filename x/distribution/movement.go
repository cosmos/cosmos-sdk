package stake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// burn burn burn
func BurnFeeHandler(ctx sdk.Context, _ sdk.Tx, collectedFees sdk.Coins) {}

//// Handle fee distribution to the validators and delegators
//func (k Keeper) FeeHandler(ctx sdk.Context, collectedFees sdk.Coins) {
//pool := k.GetPool(ctx)
//params := k.GetParams(ctx)

//// XXX determine
//candidate := NewCandidate(addrs[0], pks[0], Description{})

//// calculate the proposer reward
//precommitPower := k.GetTotalPrecommitVotingPower(ctx)
//toProposer := coinsMulRat(collectedFees, (sdk.NewRat(1, 100).Add(sdk.NewRat(4, 100).Mul(precommitPower).Quo(pool.BondedShares))))
//candidate.ProposerRewardPool = candidate.ProposerRewardPool.Plus(toProposer)

//toReservePool := coinsMulRat(collectedFees, params.ReservePoolFee)
//pool.FeeReservePool = pool.FeeReservePool.Plus(toReservePool)

//distributedReward := (collectedFees.Minus(toProposer)).Minus(toReservePool)
//pool.FeePool = pool.FeePool.Plus(distributedReward)
//pool.FeeSumReceived = pool.FeeSumReceived.Plus(distributedReward)
//pool.FeeRecent = distributedReward

//// lastly update the FeeRecent term
//pool.FeeRecent = collectedFees

//k.setPool(ctx, pool)
//}

//func coinsMulRat(coins sdk.Coins, rat sdk.Rat) sdk.Coins {
//var res sdk.Coins
//for _, coin := range coins {
//coinMulAmt := rat.Mul(sdk.NewRat(coin.Amount)).Evaluate()
//coinMul := sdk.Coins{{coin.Denom, coinMulAmt}}
//res = res.Plus(coinMul)
//}
//return res
//}

////____________________________________________________________________________-

//// calculate adjustment changes for a candidate at a height
//func CalculateAdjustmentChange(candidate Candidate, pool Pool, denoms []string, height int64) (Candidate, Pool) {

//heightRat := sdk.NewRat(height)
//lastHeightRat := sdk.NewRat(height - 1)
//candidateFeeCount := candidate.BondedShares.Mul(heightRat)
//poolFeeCount := pool.BondedShares.Mul(heightRat)

//for i, denom := range denoms {
//poolFeeSumReceived := sdk.NewRat(pool.FeeSumReceived.AmountOf(denom))
//poolFeeRecent := sdk.NewRat(pool.FeeRecent.AmountOf(denom))
//// calculate simple and projected pools
//simplePool := candidateFeeCount.Quo(poolFeeCount).Mul(poolFeeSumReceived)
//calc1 := candidate.PrevBondedShares.Mul(lastHeightRat).Quo(pool.PrevBondedShares.Mul(lastHeightRat)).Mul(poolFeeRecent)
//calc2 := candidate.BondedShares.Quo(pool.BondedShares).Mul(poolFeeRecent)
//projectedPool := calc1.Add(calc2)

//AdjustmentChange := simplePool.Sub(projectedPool)
//candidate.FeeAdjustments[i] = candidate.FeeAdjustments[i].Add(AdjustmentChange)
//pool.FeeAdjustments[i] = pool.FeeAdjustments[i].Add(AdjustmentChange)
//}

//return candidate, pool
//}
