package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// XXX TODO
func (k Keeper) AllocateFees(ctx sdk.Context, feesCollected sdk.Coins, proposerAddr sdk.ConsAddress,
	sumPowerPrecommitValidators, totalBondedTokens, communityTax, proposerCommissionRate sdk.Dec) {

	feePool := k.GetFeePool()
	proserValidator := k.stakeKeeper.GetValidatorFromConsAddr(ctx, proposerAddr)
	proposer := k.GetFeeDistribution(ctx, proserValidator.OperatorAddr)

	feesCollectedDec := NewDecCoins(feesCollected)
	proposerMultiplier := sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(4, 2).Mul(
		sumPowerPrecommitValidators).Div(totalBondedTokens))
	proposerReward := feesCollectedDec.Mul(proposerMultiplier)

	commission := proposerReward.Mul(proposerCommissionRate)
	proposer.PoolCommission = proposer.PoolCommission.Add(commission)
	proposer.Pool = proposer.Pool.Add(proposerReward.Sub(commission))

	communityFunding := feesCollectedDec.Mul(communityTax)
	feePool.CommunityFund = feePool.CommunityFund.Add(communityFunding)

	poolReceived = feesCollectedDec - proposerReward - communityFunding
	feePool.Pool = feePool.Pool.Add(poolReceived)

	SetValidatorDistribution(proposer)
	SetFeePool(feePool)
}
