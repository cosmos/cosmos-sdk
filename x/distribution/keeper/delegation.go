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

// remove a delegator distribution info
func (k Keeper) RemoveDelegatorDistInfo(ctx sdk.Context, delAddr sdk.AccAddress,
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
	validatorAddr sdk.ValAddress) {

	height := ctx.BlockHeight()
	bondedTokens := k.stakeKeeper.TotalPower(ctx)
	feePool := k.GetFeePool(ctx)
	delInfo := k.GetDelegatorDistInfo(ctx, delegatorAddr, validatorAddr)
	valInfo := k.GetValidatorDistInfo(ctx, validatorAddr)
	validator := k.stakeKeeper.GetValidator(ctx, validatorAddr)

	delInfo, feePool, withdraw := delInfo.WithdrawRewards(ctx, feePool, valInfo, height, bondedTokens,
		validator.Tokens, validator.DelegatorShares, validator.Commission)

	k.SetFeePool(ctx, feePool)
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, delegatorAddr)
	k.bankKeeper.AddCoins(ctx, withdrawAddr, withdraw.TruncateDecimal())
}

//___________________________________________________________________________________________

// return all rewards for all delegations of a delegator
func (k Keeper) WithdrawDelegationRewardsAll(ctx sdk.Context, delegatorAddr sdk.AccAddress) {
	height := ctx.BlockHeight()
	withdraw = k.GetDelegatorRewardsAll(ctx, delegatorAddr, height)
	withdrawAddr := k.GetDelegatorWithdrawAddr(delegatorAddr)
	k.coinsKeeper.AddCoins(withdrawAddr, withdraw.Amount.TruncateDecimal())
}

// return all rewards for all delegations of a delegator
func (k Keeper) GetDelegatorRewardsAll(ctx sdk.Context, delAddr sdk.AccAddress, height int64) types.DecCoins {

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
