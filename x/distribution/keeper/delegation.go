package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// check whether a delegator distribution info exists
func (k Keeper) HasDelegationDistInfo(ctx sdk.Context, delAddr sdk.AccAddress,
	valOperatorAddr sdk.ValAddress) (has bool) {
	store := ctx.KVStore(k.storeKey)
	return store.Has(GetDelegationDistInfoKey(delAddr, valOperatorAddr))
}

// get the delegator distribution info
func (k Keeper) GetDelegationDistInfo(ctx sdk.Context, delAddr sdk.AccAddress,
	valOperatorAddr sdk.ValAddress) (ddi types.DelegationDistInfo) {

	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetDelegationDistInfoKey(delAddr, valOperatorAddr))
	if b == nil {
		panic("Stored delegation-distribution info should not have been nil")
	}

	k.cdc.MustUnmarshalBinary(b, &ddi)
	return
}

// set the delegator distribution info
func (k Keeper) SetDelegationDistInfo(ctx sdk.Context, ddi types.DelegationDistInfo) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(ddi)
	store.Set(GetDelegationDistInfoKey(ddi.DelegatorAddr, ddi.ValOperatorAddr), b)
}

// remove a delegator distribution info
func (k Keeper) RemoveDelegationDistInfo(ctx sdk.Context, delAddr sdk.AccAddress,
	valOperatorAddr sdk.ValAddress) {

	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegationDistInfoKey(delAddr, valOperatorAddr))
}

//___________________________________________________________________________________________

// get the delegator withdraw address, return the delegator address if not set
func (k Keeper) GetDelegatorWithdrawAddr(ctx sdk.Context, delAddr sdk.AccAddress) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(GetDelegatorWithdrawAddrKey(delAddr))
	if b == nil {
		return delAddr
	}
	return sdk.AccAddress(b)
}

// set the delegator withdraw address
func (k Keeper) SetDelegatorWithdrawAddr(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetDelegatorWithdrawAddrKey(delAddr), withdrawAddr.Bytes())
}

// remove a delegator withdraw info
func (k Keeper) RemoveDelegatorWithdrawAddr(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegatorWithdrawAddrKey(delAddr))
}

//___________________________________________________________________________________________

// withdraw all the rewards for a single delegation
func (k Keeper) WithdrawDelegationReward(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress) sdk.Error {

	if !k.HasDelegationDistInfo(ctx, delegatorAddr, validatorAddr) {
		return types.ErrNoDelegationDistInfo(k.codespace)
	}

	height := ctx.BlockHeight()
	bondedTokens := k.stakeKeeper.TotalPower(ctx)
	feePool := k.GetFeePool(ctx)
	delInfo := k.GetDelegationDistInfo(ctx, delegatorAddr, validatorAddr)
	valInfo := k.GetValidatorDistInfo(ctx, validatorAddr)
	validator := k.stakeKeeper.Validator(ctx, validatorAddr)
	delegation := k.stakeKeeper.Delegation(ctx, delegatorAddr, validatorAddr)

	delInfo, valInfo, feePool, withdraw := delInfo.WithdrawRewards(feePool, valInfo, height, bondedTokens,
		validator.GetTokens(), validator.GetDelegatorShares(), delegation.GetShares(), validator.GetCommission())

	k.SetValidatorDistInfo(ctx, valInfo)
	k.SetDelegationDistInfo(ctx, delInfo)
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, delegatorAddr)
	coinsToAdd, change := withdraw.TruncateDecimal()
	feePool.CommunityPool = feePool.CommunityPool.Plus(change)
	k.SetFeePool(ctx, feePool)
	_, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coinsToAdd)
	if err != nil {
		panic(err)
	}
	return nil
}

//___________________________________________________________________________________________

// return all rewards for all delegations of a delegator
func (k Keeper) WithdrawDelegationRewardsAll(ctx sdk.Context, delegatorAddr sdk.AccAddress) {
	height := ctx.BlockHeight()
	withdraw := k.getDelegatorRewardsAll(ctx, delegatorAddr, height)
	feePool := k.GetFeePool(ctx)
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, delegatorAddr)
	coinsToAdd, change := withdraw.TruncateDecimal()
	feePool.CommunityPool = feePool.CommunityPool.Plus(change)
	k.SetFeePool(ctx, feePool)
	_, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coinsToAdd)
	if err != nil {
		panic(err)
	}
}

// return all rewards for all delegations of a delegator
func (k Keeper) getDelegatorRewardsAll(ctx sdk.Context, delAddr sdk.AccAddress, height int64) types.DecCoins {

	withdraw := types.DecCoins{}
	bondedTokens := k.stakeKeeper.TotalPower(ctx)

	// iterate over all the delegations
	operationAtDelegation := func(_ int64, del sdk.Delegation) (stop bool) {
		feePool := k.GetFeePool(ctx)
		valAddr := del.GetValidator()
		delInfo := k.GetDelegationDistInfo(ctx, delAddr, valAddr)
		valInfo := k.GetValidatorDistInfo(ctx, valAddr)
		validator := k.stakeKeeper.Validator(ctx, valAddr)
		delegation := k.stakeKeeper.Delegation(ctx, delAddr, valAddr)

		delInfo, valInfo, feePool, diWithdraw := delInfo.WithdrawRewards(feePool, valInfo, height, bondedTokens,
			validator.GetTokens(), validator.GetDelegatorShares(), delegation.GetShares(), validator.GetCommission())
		withdraw = withdraw.Plus(diWithdraw)
		k.SetFeePool(ctx, feePool)
		k.SetValidatorDistInfo(ctx, valInfo)
		k.SetDelegationDistInfo(ctx, delInfo)
		return false
	}
	k.stakeKeeper.IterateDelegations(ctx, delAddr, operationAtDelegation)
	return withdraw
}
