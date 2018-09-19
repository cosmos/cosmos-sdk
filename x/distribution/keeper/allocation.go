package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Allocate fees handles distribution of the collected fees
func (k Keeper) AllocateFees(ctx sdk.Context) {

	// get the proposer of this block
	proposerConsAddr := k.GetProposerConsAddr(ctx)
	proserValidator := k.stakeKeeper.GetValidatorFromConsAddr(ctx, proposerConsAddr)
	proposerDist := k.GetFeeDistribution(ctx, proserValidator.OperatorAddr)

	// get the fees which have been getting collected through all the
	// transactions in the block
	feesCollected := k.FeeCollectionKeeper.GetCollectedFees(ctx)
	feesCollectedDec := NewDecCoins(feesCollected)

	// allocated rewards to proposer
	stakePool := k.stakeKeeper.GetPool(ctx)
	sumPowerPrecommitValidators := sdk.NewDec(1) // XXX TODO actually calculate this
	proposerMultiplier := sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(4, 2).Mul(
		sumPowerPrecommitValidators).Div(stakePool.BondedTokens))
	proposerReward := feesCollectedDec.Mul(proposerMultiplier)

	// apply commission
	commission := proposerReward.Mul(proserValidator.Commission)
	proposerDist.PoolCommission = proposerDist.PoolCommission.Add(commission)
	proposerDist.Pool = proposerDist.Pool.Add(proposerReward.Sub(commission))

	// allocate community funding
	communityTax := k.GetCommunityTax(ctx)
	communityFunding := feesCollectedDec.Mul(communityTax)
	feePool := k.GetFeePool(ctx)
	feePool.CommunityFund = feePool.CommunityFund.Add(communityFunding)

	// set the global pool within the distribution module
	poolReceived := feesCollectedDec.Sub(proposerReward).Sub(communityFunding)
	feePool.Pool = feePool.Pool.Add(poolReceived)

	SetValidatorDistribution(proposerDist)
	SetFeePool(feePool)

	// clear the now distributed fees
	k.FeeCollectionKeeper.ClearCollectedFees(ctx)
}
