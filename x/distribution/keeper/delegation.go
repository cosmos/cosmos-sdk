package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
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

// withdraw all the rewards for a single delegation
func (k Keeper) WithdrawDelegationReward(ctx sdk.Context, delegatorAddr,
	withdrawAddr sdk.AccAddress, validatorAddr sdk.ValAddress) {

	height := ctx.BlockHeight()
	pool := k.sk.GetPool(ctx)
	feePool := k.GetFeePool(ctx)
	delInfo := k.GetDelegationDistInfo(ctx, delegatorAddr, validatorAddr)
	valInfo := k.GetValidatorDistInfo(ctx, validatorAddr)
	validator := k.GetValidator(ctx, validatorAddr)

	feePool, withdraw := delInfo.WithdrawRewards(feePool, valInfo, height, pool.BondedTokens,
		validator.Tokens, validator.DelegatorShares, validator.Commission)

	k.SetFeePool(ctx, feePool)
	k.ck.AddCoins(ctx, withdrawAddr, withdraw.TruncateDecimal())
}

///////////////////////////////////////////////////////////////////////////////////////

// return all rewards for all delegations of a delegator
func (k Keeper) WithdrawDelegationRewardsAll(ctx sdk.Context, delegatorAddr, withdrawAddr sdk.AccAddress) {
	height := ctx.BlockHeight()
	withdraw = GetDelegatorRewardsAll(ctx, delegatorAddr, height)
	k.coinsKeeper.AddCoins(withdrawAddr, withdraw.Amount.TruncateDecimal())
}

// return all rewards for all delegations of a delegator
func (k Keeper) GetDelegatorRewardsAll(ctx sdk.Context, delAddr sdk.AccAddress, height int64) DecCoins {

	withdraw := sdk.NewDec(0)
	pool := k.sk.GetPool(ctx)
	feePool := k.GetFeePool(ctx)

	// iterate over all the delegations
	operationAtDelegation := func(_ int64, del types.Delegation) (stop bool) {
		delInfo := k.GetDelegationDistInfo(ctx, delAddr, del.ValidatorAddr)
		valInfo := k.GetValidatorDistInfo(ctx, del.ValidatorAddr)
		validator := k.sk.GetValidator(ctx, del.ValidatorAddr)

		feePool, diWithdraw := delInfo.WithdrawRewards(feePool, valInfo, height, pool.BondedTokens,
			validator.Tokens, validator.DelegatorShares, validator.Commission)
		withdraw = withdraw.Add(diWithdraw)
		SetFeePool(feePool)
		return false
	}
	k.stakeKeeper.IterateDelegations(ctx, delAddr, operationAtDelegation)

	k.SetFeePool(ctx, feePool)
	return withdraw
}
