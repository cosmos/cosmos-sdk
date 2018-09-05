package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// XXX TODO
func (k Keeper) AllocateFees(ctx sdk.Context, feesCollected sdk.Coins, proposerAddr sdk.ConsAddrs,
	sumPowerPrecommitValidators, totalBondedTokens, communityTax, proposerCommissionRate sdk.Dec) {

	feePool := k.GetFeePool()
	validator := k.stakeKeeper.GetValidatorFromConsAddr(ctx, 
	proposerOpAddr := Stake.GetValidator
	proposer := k.GetFeeDistribution(ctx, proposerOpAddr)

	feesCollectedDec = MakeDecCoins(feesCollected)
	proposerReward = feesCollectedDec.Mul(sdk.NewDecWithPrec(1, 2) + sdk.NewDecWithPrec(1, 2).Mul(sumPowerPrecommitValidators)/totalBondedTokens)

	commission = proposerReward * proposerCommissionRate
	proposer.PoolCommission += commission
	proposer.Pool += proposerReward - commission

	communityFunding = feesCollectedDec * communityTax
	feePool.CommunityFund += communityFunding

	poolReceived = feesCollectedDec - proposerReward - communityFunding
	feePool.Pool += poolReceived
	feePool.EverReceivedPool += poolReceived
	feePool.LastReceivedPool = poolReceived

	SetValidatorDistribution(proposer)
	SetFeePool(feePool)
}
