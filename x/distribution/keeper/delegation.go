package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// get the delegator distribution info
func (k Keeper) GetDelegatorDistInfo(ctx sdk.Context, delAddr sdk.AccAddress,
	valOperatorAddr sdk.ValAddress) (ddi types.DelegatorDistInfo) {

	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetDelegationDistInfoKey(delAddr, valOperatorAddr))
	if b == nil {
		panic("Stored delegation-distribution info should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &ddi)
	return
}

// set the delegator distribution info
func (k Keeper) SetDelegatorDistInfo(ctx sdk.Context, ddi types.DelegatorDistInfo) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(ddi)
	store.Set(GetDelegationDistInfoKey(ddi.DelegatorAddr, ddi.ValOperatorAddr), b)
}

//___________________________________________________________________________________________

// XXX TODO
func (k Keeper) WithdrawDelegationReward(ctx sdk.Context, delegatorAddr, validatorAddr, withdrawAddr sdk.AccAddress) {
	height = ctx.BlockHeight()

	// get all distribution scenarios
	pool = stake.GetPool()
	feePool = GetFeePool()
	delInfo = GetDelegationDistInfo(delegatorAddr,
		validatorAddr)
	valInfo = GetValidatorDistInfo(validatorAddr)
	validator = GetValidator(validatorAddr)

	feePool, withdraw = delInfo.WithdrawRewards(feePool, valInfo, height, pool.BondedTokens,
		validator.Tokens, validator.DelegatorShares, validator.Commission)

	SetFeePool(feePool)
	AddCoins(withdrawAddr, withdraw.TruncateDecimal())
}

///////////////////////////////////////////////////////////////////////////////////////

// XXX TODO
func (k Keeper) WithdrawDelegationRewardsAll(ctx sdk.Context, delegatorAddr, withdrawAddr sdk.AccAddress) {
	height = ctx.BlockHeight()
	withdraw = GetDelegatorRewardsAll(delegatorAddr, height)
	k.coinsKeeper.AddCoins(withdrawAddr, withdraw.Amount.TruncateDecimal())
}

// XXX TODO
func (k Keeper) GetDelegatorRewardsAll(ctx sdk.Context, delAddr sdk.AccAddress, height int64) DecCoins {

	// collect all entitled rewards
	withdraw = 0
	pool = stake.GetPool()
	feePool = GetFeePool()

	// iterate over all the delegations
	operationAtDelegation := func(_ int64, del types.Delegation) (stop bool) {
		delInfo = GetDelegationDistInfo(delAddr, del.ValidatorAddr)
		valInfo = GetValidatorDistInfo(del.ValidatorAddr)
		validator = GetValidator(del.ValidatorAddr)

		feePool, diWithdraw = delInfo.WithdrawRewards(feePool, valInfo, height, pool.BondedTokens,
			validator.Tokens, validator.DelegatorShares, validator.Commission)
		withdraw += diWithdraw
		return false
	}
	k.stakeKeeper.IterateDelegations(ctx, delAddr, operationAtDelegation)

	SetFeePool(feePool)
	return withdraw
}
