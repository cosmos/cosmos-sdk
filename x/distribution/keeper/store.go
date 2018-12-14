package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// get the global fee pool distribution info
func (k Keeper) GetFeePool(ctx sdk.Context) (feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(FeePoolKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &feePool)
	return
}

// set the global fee pool distribution info
func (k Keeper) SetFeePool(ctx sdk.Context, feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(feePool)
	store.Set(FeePoolKey, b)
}

// get the proposer public key for this block
func (k Keeper) GetPreviousProposerConsAddr(ctx sdk.Context) (consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(ProposerKey)
	if b == nil {
		panic("Previous proposer not set")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &consAddr)
	return
}

// set the proposer public key for this block
func (k Keeper) SetPreviousProposerConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(consAddr)
	store.Set(ProposerKey, b)
}

// get the starting period associated with a delegator
func (k Keeper) GetDelegatorStartingPeriod(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) (period types.DelegatorStartingPeriod) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetDelegatorStartingPeriodKey(val, del))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &period)
	return
}

// set the starting period associated with a delegator
func (k Keeper) setDelegatorStartingPeriod(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress, period types.DelegatorStartingPeriod) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(period)
	store.Set(GetDelegatorStartingPeriodKey(val, del), b)
}

// get historical rewards for a particular period
func (k Keeper) GetValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress, period uint64) (rewards types.ValidatorHistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorHistoricalRewardsKey(val, period))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &rewards)
	return
}

// set historical rewards for a particular period
func (k Keeper) setValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(rewards)
	store.Set(GetValidatorHistoricalRewardsKey(val, period), b)
}

// get current rewards for a validator
func (k Keeper) GetValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress) (rewards types.ValidatorCurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorCurrentRewardsKey(val))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &rewards)
	return
}

// set current rewards for a validator
func (k Keeper) setValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress, rewards types.ValidatorCurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(rewards)
	store.Set(GetValidatorCurrentRewardsKey(val), b)
}

// get accumulated commission for a validator
func (k Keeper) GetValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress) (commission types.ValidatorAccumulatedCommission) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorAccumulatedCommissionKey(val))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &commission)
	return
}

// set accumulated commission for a validator
func (k Keeper) setValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(commission)
	store.Set(GetValidatorAccumulatedCommissionKey(val), b)
}

// get outstanding rewards
func (k Keeper) GetOutstandingRewards(ctx sdk.Context) (rewards types.OutstandingRewards) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(OutstandingRewardsKey)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &rewards)
	return
}

// set outstanding rewards
func (k Keeper) SetOutstandingRewards(ctx sdk.Context, rewards types.OutstandingRewards) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(rewards)
	store.Set(OutstandingRewardsKey, b)
}
