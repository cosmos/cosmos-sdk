package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Allocate fees handles distribution of the collected fees
func (k Keeper) AllocateFees(ctx sdk.Context) {
	ctx.Logger().With("module", "x/distribution").Error(fmt.Sprintf("allocation height: %v", ctx.BlockHeight()))

	// if there is no power in the system nothing should be allocated
	bondedTokens := k.stakeKeeper.TotalPower(ctx).TruncateInt()
	if bondedTokens.IsZero() {
		return
	}

	// get the proposer of this block
	proposerConsAddr := k.GetProposerConsAddr(ctx)
	proposerValidator := k.stakeKeeper.ValidatorByConsAddr(ctx, proposerConsAddr)
	proposerDist := k.GetValidatorDistInfo(ctx, proposerValidator.GetOperator())

	// get the fees which have been getting collected through all the
	// transactions in the block
	feesCollected := k.feeCollectionKeeper.GetCollectedFees(ctx)
	feesCollectedDec := types.NewDecCoins(feesCollected)

	// allocated rewards to proposer
	sumPowerPrecommitValidators := k.GetSumPrecommitPower(ctx)
	percentVoting := sdk.NewDec(sumPowerPrecommitValidators).QuoInt(bondedTokens)

	// rare edge case for rounding tendermint power vs bonded decimal power
	if percentVoting.GT(sdk.OneDec()) {
		percentVoting = sdk.OneDec()
	}

	proposerMultiplier := sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(4, 2).Mul(percentVoting))
	proposerReward := feesCollectedDec.MulDec(proposerMultiplier)

	// apply commission
	commission := proposerReward.MulDec(proposerValidator.GetCommission())
	remaining := proposerReward.MulDec(sdk.OneDec().Sub(proposerValidator.GetCommission()))
	proposerDist.PoolCommission = proposerDist.PoolCommission.Plus(commission)
	proposerDist.Pool = proposerDist.Pool.Plus(remaining)

	// allocate community funding
	communityTax := k.GetCommunityTax(ctx)
	communityFunding := feesCollectedDec.MulDec(communityTax)
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Plus(communityFunding)

	// set the global pool within the distribution module
	poolReceived := feesCollectedDec.Minus(proposerReward).Minus(communityFunding)
	feePool.Pool = feePool.Pool.Plus(poolReceived)

	k.SetValidatorDistInfo(ctx, proposerDist)
	k.SetFeePool(ctx, feePool)

	// clear the now distributed fees
	k.feeCollectionKeeper.ClearCollectedFees(ctx)
}
