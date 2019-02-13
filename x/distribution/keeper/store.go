package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// get the delegator withdraw address, defaulting to the delegator address
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

// delete a delegator withdraw addr
func (k Keeper) DeleteDelegatorWithdrawAddr(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegatorWithdrawAddrKey(delAddr))
}

// iterate over delegator withdraw addrs
func (k Keeper) IterateDelegatorWithdrawAddrs(ctx sdk.Context, handler func(del sdk.AccAddress, addr sdk.AccAddress) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, DelegatorWithdrawAddrPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addr := sdk.AccAddress(iter.Value())
		del := GetDelegatorWithdrawInfoAddress(iter.Key())
		if handler(del, addr) {
			break
		}
	}
}

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

// get the starting info associated with a delegator
func (k Keeper) GetDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) (period types.DelegatorStartingInfo) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetDelegatorStartingInfoKey(val, del))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &period)
	return
}

// set the starting info associated with a delegator
func (k Keeper) SetDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress, period types.DelegatorStartingInfo) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(period)
	store.Set(GetDelegatorStartingInfoKey(val, del), b)
}

// check existence of the starting info associated with a delegator
func (k Keeper) HasDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(GetDelegatorStartingInfoKey(val, del))
}

// delete the starting info associated with a delegator
func (k Keeper) DeleteDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetDelegatorStartingInfoKey(val, del))
}

// iterate over delegator starting infos
func (k Keeper) IterateDelegatorStartingInfos(ctx sdk.Context, handler func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, DelegatorStartingInfoPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var info types.DelegatorStartingInfo
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &info)
		val, del := GetDelegatorStartingInfoAddresses(iter.Key())
		if handler(val, del, info) {
			break
		}
	}
}

// get historical rewards for a particular period
func (k Keeper) GetValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress, period uint64) (rewards types.ValidatorHistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorHistoricalRewardsKey(val, period))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &rewards)
	return
}

// set historical rewards for a particular period
func (k Keeper) SetValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(rewards)
	store.Set(GetValidatorHistoricalRewardsKey(val, period), b)
}

// iterate over historical rewards
func (k Keeper) IterateValidatorHistoricalRewards(ctx sdk.Context, handler func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorHistoricalRewards
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &rewards)
		addr, period := GetValidatorHistoricalRewardsAddressPeriod(iter.Key())
		if handler(addr, period, rewards) {
			break
		}
	}
}

// delete a historical reward
func (k Keeper) DeleteValidatorHistoricalReward(ctx sdk.Context, val sdk.ValAddress, period uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetValidatorHistoricalRewardsKey(val, period))
}

// delete historical rewards for a validator
func (k Keeper) DeleteValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, GetValidatorHistoricalRewardsPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// delete all historical rewards
func (k Keeper) DeleteAllValidatorHistoricalRewards(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// historical reference count (used for testcases)
func (k Keeper) GetValidatorHistoricalReferenceCount(ctx sdk.Context) (count uint64) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorHistoricalRewards
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &rewards)
		count += uint64(rewards.ReferenceCount)
	}
	return
}

// get current rewards for a validator
func (k Keeper) GetValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress) (rewards types.ValidatorCurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorCurrentRewardsKey(val))
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &rewards)
	return
}

// set current rewards for a validator
func (k Keeper) SetValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress, rewards types.ValidatorCurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(rewards)
	store.Set(GetValidatorCurrentRewardsKey(val), b)
}

// delete current rewards for a validator
func (k Keeper) DeleteValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetValidatorCurrentRewardsKey(val))
}

// iterate over current rewards
func (k Keeper) IterateValidatorCurrentRewards(ctx sdk.Context, handler func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorCurrentRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorCurrentRewards
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &rewards)
		addr := GetValidatorCurrentRewardsAddress(iter.Key())
		if handler(addr, rewards) {
			break
		}
	}
}

// get accumulated commission for a validator
func (k Keeper) GetValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress) (commission types.ValidatorAccumulatedCommission) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorAccumulatedCommissionKey(val))
	if b == nil {
		return types.ValidatorAccumulatedCommission{}
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &commission)
	return
}

// set accumulated commission for a validator
func (k Keeper) SetValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(commission)
	store.Set(GetValidatorAccumulatedCommissionKey(val), b)
}

// delete accumulated commission for a validator
func (k Keeper) DeleteValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetValidatorAccumulatedCommissionKey(val))
}

// iterate over accumulated commissions
func (k Keeper) IterateValidatorAccumulatedCommissions(ctx sdk.Context, handler func(val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorAccumulatedCommissionPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var commission types.ValidatorAccumulatedCommission
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &commission)
		addr := GetValidatorAccumulatedCommissionAddress(iter.Key())
		if handler(addr, commission) {
			break
		}
	}
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

// get slash event for height
func (k Keeper) GetValidatorSlashEvent(ctx sdk.Context, val sdk.ValAddress, height uint64) (event types.ValidatorSlashEvent, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetValidatorSlashEventKey(val, height))
	if b == nil {
		return types.ValidatorSlashEvent{}, false
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &event)
	return event, true
}

// set slash event for height
func (k Keeper) SetValidatorSlashEvent(ctx sdk.Context, val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(event)
	store.Set(GetValidatorSlashEventKey(val, height), b)
}

// iterate over slash events between heights, inclusive
func (k Keeper) IterateValidatorSlashEventsBetween(ctx sdk.Context, val sdk.ValAddress, startingHeight uint64, endingHeight uint64,
	handler func(height uint64, event types.ValidatorSlashEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := store.Iterator(
		GetValidatorSlashEventKey(val, startingHeight),
		GetValidatorSlashEventKey(val, endingHeight+1),
	)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var event types.ValidatorSlashEvent
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &event)
		_, height := GetValidatorSlashEventAddressHeight(iter.Key())
		if handler(height, event) {
			break
		}
	}
}

// iterate over all slash events
func (k Keeper) IterateValidatorSlashEvents(ctx sdk.Context, handler func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorSlashEventPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var event types.ValidatorSlashEvent
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &event)
		val, height := GetValidatorSlashEventAddressHeight(iter.Key())
		if handler(val, height, event) {
			break
		}
	}
}

// delete slash events for a particular validator
func (k Keeper) DeleteValidatorSlashEvents(ctx sdk.Context, val sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, GetValidatorSlashEventPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// delete all slash events
func (k Keeper) DeleteAllValidatorSlashEvents(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, ValidatorSlashEventPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
