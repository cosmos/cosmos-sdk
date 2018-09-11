package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// Allocate fees handles distribution of the collected fees
func (k Keeper) AllocateFees(ctx sdk.Context, _ sdk.Tx, feesCollected sdk.Coins, proposerAddr sdk.ConsAddress) {

	sumPowerPrecommitValidators := sdk.NewDec(1) // XXX TODO actually calculate this
	communityTax := sdk.NewDecWithPrec(1, 2)     // XXX TODO get from global params store

	feePool := k.GetFeePool(ctx)
	stakePool := k.stakeKeeper.GetPool(ctx)

	proserValidator := k.stakeKeeper.GetValidatorFromConsAddr(ctx, proposerAddr)
	proposerDist := k.GetFeeDistribution(ctx, proserValidator.OperatorAddr)

	feesCollectedDec := NewDecCoins(feesCollected)
	proposerMultiplier := sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(4, 2).Mul(
		sumPowerPrecommitValidators).Div(stakePool.BondedTokens))
	proposerReward := feesCollectedDec.Mul(proposerMultiplier)

	commission := proposerReward.Mul(proserValidator.Commission)
	proposerDist.PoolCommission = proposerDist.PoolCommission.Add(commission)
	proposerDist.Pool = proposerDist.Pool.Add(proposerReward.Sub(commission))

	communityFunding := feesCollectedDec.Mul(communityTax)
	feePool.CommunityFund = feePool.CommunityFund.Add(communityFunding)

	poolReceived := feesCollectedDec.Sub(proposerReward).Sub(communityFunding)
	feePool.Pool = feePool.Pool.Add(poolReceived)

	SetValidatorDistribution(proposerDist)
	SetFeePool(feePool)
}
