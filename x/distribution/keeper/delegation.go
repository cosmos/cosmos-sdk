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

// return all rewards for a delegation
func (k Keeper) withdrawDelegationReward(ctx sdk.Context, delAddr sdk.AccAddress,
	height int64, lastTotalPower sdk.Dec) (
	feePool, ValidatorDistInfo, DelegatorDistInfo, types.DecCoins) {

	valAddr := del.GetValidatorAddr()
	wc := k.GetWithdrawContext(ctx, valAddr)

	delInfo := k.GetDelegationDistInfo(ctx, delAddr, valAddr)
	validator := k.stakeKeeper.Validator(ctx, valAddr)
	delegation := k.stakeKeeper.Delegation(ctx, delAddr, valAddr)

	delInfo, valInfo, feePool, withdraw := delInfo.WithdrawRewards(wc,
		validator.GetDelegatorShares(), delegation.GetShares())

	return feePool, valInfo, delInfo, withdraw
}

// estimate all rewards for all delegations of a delegator
func (k Keeper) estimateDelegationReward(ctx sdk.Context, delAddr sdk.AccAddress,
	height int64, lastTotalPower sdk.Dec) types.DecCoins {

	valAddr := del.GetValidatorAddr()
	wc := k.GetWithdrawContext(ctx, valAddr)

	delInfo := k.GetDelegationDistInfo(ctx, delAddr, valAddr)
	validator := k.stakeKeeper.Validator(ctx, valAddr)
	delegation := k.stakeKeeper.Delegation(ctx, delAddr, valAddr)

	estimation := delInfo.EstimateRewards(wc,
		validator.GetDelegatorShares(), delegation.GetShares())

	return estimation
}

//___________________________________________________________________________________________

// withdraw all rewards for a single delegation
// NOTE: This gets called "onDelegationSharesModified",
// meaning any changes to bonded coins
func (k Keeper) WithdrawDelegationReward(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	valAddr sdk.ValAddress) sdk.Error {

	if !k.HasDelegationDistInfo(ctx, delegatorAddr, valAddr) {
		return types.ErrNoDelegationDistInfo(k.codespace)
	}

	feePool, valInfo, delInfo, withdraw :=
		withdrawDelegationReward(ctx, delAddr, height, lastTotalPower)

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

// estimate rewards for a single delegation
func (k Keeper) EstimateDelegationReward(ctx sdk.Context, delegatorAddr sdk.AccAddress,
	valAddr sdk.ValAddress) (sdk.Coins, sdk.Error) {

	if !k.HasDelegationDistInfo(ctx, delegatorAddr, valAddr) {
		return types.ErrNoDelegationDistInfo(k.codespace)
	}
	estCoins := estimateDelegationReward(ctx, delAddr, height, lastTotalPower)
	return estCoins.TruncateDecimal()
}

//___________________________________________________________________________________________

// return all rewards for all delegations of a delegator
func (k Keeper) WithdrawDelegationRewardsAll(ctx sdk.Context, delegatorAddr sdk.AccAddress) {
	height := ctx.BlockHeight()

	// iterate over all the delegations
	withdraw := types.DecCoins{}
	lastTotalPower := sdk.NewDecFromInt(k.stakeKeeper.GetLastTotalPower(ctx))
	operationAtDelegation := func(_ int64, del sdk.Delegation) (stop bool) {

		feePool, valInfo, delInfo, diWithdraw :=
			withdrawDelegationReward(ctx, delAddr, height, lastTotalPower)
		withdraw = withdraw.Plus(diWithdraw)
		k.SetFeePool(ctx, feePool)
		k.SetValidatorDistInfo(ctx, valInfo)
		k.SetDelegationDistInfo(ctx, delInfo)
		return false
	}
	k.stakeKeeper.IterateDelegations(ctx, delAddr, operationAtDelegation)

	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, delegatorAddr)
	coinsToAdd, change := withdraw.TruncateDecimal()
	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Plus(change)
	k.SetFeePool(ctx, feePool)
	_, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coinsToAdd)
	if err != nil {
		panic(err)
	}
}

// estimate all rewards for all delegations of a delegator
func (k Keeper) EstimateDelegationRewardsAll(ctx sdk.Context,
	delegatorAddr sdk.AccAddress) sdk.Coins {

	height := ctx.BlockHeight()

	// iterate over all the delegations
	total := types.DecCoins{}
	lastTotalPower := sdk.NewDecFromInt(k.stakeKeeper.GetLastTotalPower(ctx))
	operationAtDelegation := func(_ int64, del sdk.Delegation) (stop bool) {
		est := estimateDelegationReward(ctx, delAddr, height, lastTotalPower)
		total = total.Plus(est)
		return false
	}
	k.stakeKeeper.IterateDelegations(ctx, delAddr, operationAtDelegation)
	estCoins, _ := total.TruncateDecimal()
	return estCoins
}
