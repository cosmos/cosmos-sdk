package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// XXX TODO
func AllocateFees(feesCollected sdk.Coins, feePool FeePool, proposer ValidatorDistribution,
	sumPowerPrecommitValidators, totalBondedTokens, communityTax,
	proposerCommissionRate sdk.Dec) {

	feesCollectedDec = MakeDecCoins(feesCollected)
	proposerReward = feesCollectedDec * (0.01 + 0.04*sumPowerPrecommitValidators/totalBondedTokens)

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
