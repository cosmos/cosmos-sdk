package keeper

import (
	gogotypes "github.com/cosmos/gogoproto/types"

	store2 "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func decodeAddr(bz []byte) (sdk.AccAddress, error) {
	if bz == nil {
		return sdk.AccAddress{}, nil
	}

	return sdk.AccAddress(bz), nil
}

// get the delegator withdraw address, defaulting to the delegator address
func (k Keeper) GetDelegatorWithdrawAddr(ctx sdk.Context, delAddr sdk.AccAddress) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	addr, err := store2.GetAndDecode(store, decodeAddr, types.GetDelegatorWithdrawAddrKey(delAddr))
	if addr == nil {
		return delAddr
	}
	if err != nil {
		panic(err)
	}

	return addr
}

// set the delegator withdraw address
func (k Keeper) SetDelegatorWithdrawAddr(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress) {
	store := k.getStore(ctx)
	store.Set(types.GetDelegatorWithdrawAddrKey(delAddr), withdrawAddr.Bytes())
}

// delete a delegator withdraw addr
func (k Keeper) DeleteDelegatorWithdrawAddr(ctx sdk.Context, delAddr, withdrawAddr sdk.AccAddress) {
	store := k.getStore(ctx)
	store.Delete(types.GetDelegatorWithdrawAddrKey(delAddr))
}

// iterate over delegator withdraw addrs
func (k Keeper) IterateDelegatorWithdrawAddrs(ctx sdk.Context, handler func(del sdk.AccAddress, addr sdk.AccAddress) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DelegatorWithdrawAddrPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addr := sdk.AccAddress(iter.Value())
		del := types.GetDelegatorWithdrawInfoAddress(iter.Key())
		if handler(del, addr) {
			break
		}
	}
}

func (k Keeper) decodeFeePool(bz []byte) (types.FeePool, error) {
	var feePool types.FeePool
	if bz == nil {
		panic("Stored fee pool should not have been nil")
	}
	k.cdc.MustUnmarshal(bz, &feePool)
	return feePool, nil
}

// get the global fee pool distribution info
func (k Keeper) GetFeePool(ctx sdk.Context) (feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)
	feePool, err := store2.GetAndDecode(store, k.decodeFeePool, types.FeePoolKey)
	if err != nil {
		panic(err)
	}
	return
}

// set the global fee pool distribution info
func (k Keeper) SetFeePool(ctx sdk.Context, feePool types.FeePool) {
	store := k.getStore(ctx)
	b := k.cdc.MustMarshal(&feePool)
	store.Set(types.FeePoolKey, b)
}

func (k Keeper) decodeKey(bz []byte) (sdk.ConsAddress, error) {
	if bz == nil {
		panic("previous proposer not set")
	}
	addrValue := gogotypes.BytesValue{}
	k.cdc.MustUnmarshal(bz, &addrValue)
	return addrValue.GetValue(), nil
}

// GetPreviousProposerConsAddr returns the proposer consensus address for the
// current block.
func (k Keeper) GetPreviousProposerConsAddr(ctx sdk.Context) sdk.ConsAddress {
	store := ctx.KVStore(k.storeKey)
	addrValue, err := store2.GetAndDecode(store, k.decodeKey, types.ProposerKey)
	if err != nil {
		panic(err)
	}
	return addrValue
}

// set the proposer public key for this block
func (k Keeper) SetPreviousProposerConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) {
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(&gogotypes.BytesValue{Value: consAddr})
	store.Set(types.ProposerKey, bz)
}

func (k Keeper) decodePeriod(bz []byte) (types.DelegatorStartingInfo, error) {
	var period types.DelegatorStartingInfo
	if bz == nil {
		return types.DelegatorStartingInfo{}, nil
	}
	k.cdc.MustUnmarshal(bz, &period)
	return period, nil
}

// get the starting info associated with a delegator
func (k Keeper) GetDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) (period types.DelegatorStartingInfo) {
	store := ctx.KVStore(k.storeKey)
	period, err := store2.GetAndDecode(store, k.decodePeriod, types.GetDelegatorStartingInfoKey(val, del))
	if err != nil {
		panic(err)
	}
	return
}

// set the starting info associated with a delegator
func (k Keeper) SetDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress, period types.DelegatorStartingInfo) {
	store := k.getStore(ctx)
	b := k.cdc.MustMarshal(&period)
	store.Set(types.GetDelegatorStartingInfoKey(val, del), b)
}

// check existence of the starting info associated with a delegator
func (k Keeper) HasDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetDelegatorStartingInfoKey(val, del))
}

// delete the starting info associated with a delegator
func (k Keeper) DeleteDelegatorStartingInfo(ctx sdk.Context, val sdk.ValAddress, del sdk.AccAddress) {
	store := k.getStore(ctx)
	store.Delete(types.GetDelegatorStartingInfoKey(val, del))
}

// iterate over delegator starting infos
func (k Keeper) IterateDelegatorStartingInfos(ctx sdk.Context, handler func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.DelegatorStartingInfoPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var info types.DelegatorStartingInfo
		k.cdc.MustUnmarshal(iter.Value(), &info)
		val, del := types.GetDelegatorStartingInfoAddresses(iter.Key())
		if handler(val, del, info) {
			break
		}
	}
}

func (k Keeper) decodeValHistoricalRewards(bz []byte) (types.ValidatorHistoricalRewards, error) {
	var rewards types.ValidatorHistoricalRewards
	if bz == nil {
		return types.ValidatorHistoricalRewards{}, nil
	}
	k.cdc.MustUnmarshal(bz, &rewards)
	return rewards, nil
}

// get historical rewards for a particular period
func (k Keeper) GetValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress, period uint64) (rewards types.ValidatorHistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	r, err := store2.GetAndDecode(store, k.decodeValHistoricalRewards, types.GetValidatorHistoricalRewardsKey(val, period))
	if err != nil {
		panic(err)
	}
	rewards = types.ValidatorHistoricalRewards(r)
	return
}

// set historical rewards for a particular period
func (k Keeper) SetValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) {
	store := k.getStore(ctx)
	b := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetValidatorHistoricalRewardsKey(val, period), b)
}

// iterate over historical rewards
func (k Keeper) IterateValidatorHistoricalRewards(ctx sdk.Context, handler func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorHistoricalRewards
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		addr, period := types.GetValidatorHistoricalRewardsAddressPeriod(iter.Key())
		if handler(addr, period, rewards) {
			break
		}
	}
}

// delete a historical reward
func (k Keeper) DeleteValidatorHistoricalReward(ctx sdk.Context, val sdk.ValAddress, period uint64) {
	store := k.getStore(ctx)
	store.Delete(types.GetValidatorHistoricalRewardsKey(val, period))
}

// delete historical rewards for a validator
func (k Keeper) DeleteValidatorHistoricalRewards(ctx sdk.Context, val sdk.ValAddress) {
	store := k.getStore(ctx)
	iter := sdk.KVStorePrefixIterator(store, types.GetValidatorHistoricalRewardsPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// delete all historical rewards
func (k Keeper) DeleteAllValidatorHistoricalRewards(ctx sdk.Context) {
	store := k.getStore(ctx)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// historical reference count (used for testcases)
func (k Keeper) GetValidatorHistoricalReferenceCount(ctx sdk.Context) (count uint64) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorHistoricalRewards
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		count += uint64(rewards.ReferenceCount)
	}
	return
}

func (k Keeper) decodeValCurrentRewards(bz []byte) (types.ValidatorCurrentRewards, error) {
	var rewards types.ValidatorCurrentRewards
	if bz == nil {
		return types.ValidatorCurrentRewards{}, nil
	}
	k.cdc.MustUnmarshal(bz, &rewards)
	return rewards, nil
}

// get current rewards for a validator
func (k Keeper) GetValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress) (rewards types.ValidatorCurrentRewards) {
	store := ctx.KVStore(k.storeKey)
	rewards, err := store2.GetAndDecode(store, k.decodeValCurrentRewards, types.GetValidatorCurrentRewardsKey(val))
	if err != nil {
		panic(err)
	}
	return
}

// set current rewards for a validator
func (k Keeper) SetValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress, rewards types.ValidatorCurrentRewards) {
	store := k.getStore(ctx)
	b := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetValidatorCurrentRewardsKey(val), b)
}

// delete current rewards for a validator
func (k Keeper) DeleteValidatorCurrentRewards(ctx sdk.Context, val sdk.ValAddress) {
	store := k.getStore(ctx)
	store.Delete(types.GetValidatorCurrentRewardsKey(val))
}

// iterate over current rewards
func (k Keeper) IterateValidatorCurrentRewards(ctx sdk.Context, handler func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorCurrentRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorCurrentRewards
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		addr := types.GetValidatorCurrentRewardsAddress(iter.Key())
		if handler(addr, rewards) {
			break
		}
	}
}

func (k Keeper) decodeCommission(bz []byte) (types.ValidatorAccumulatedCommission, error) {
	var commission types.ValidatorAccumulatedCommission
	if bz == nil {
		return types.ValidatorAccumulatedCommission{}, nil
	}
	k.cdc.MustUnmarshal(bz, &commission)
	return commission, nil
}

// get accumulated commission for a validator
func (k Keeper) GetValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress) (commission types.ValidatorAccumulatedCommission) {
	store := ctx.KVStore(k.storeKey)
	commission, err := store2.GetAndDecode(store, k.decodeCommission, types.GetValidatorAccumulatedCommissionKey(val))
	if err != nil {
		panic(err)
	}
	return
}

// set accumulated commission for a validator
func (k Keeper) SetValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) {
	var bz []byte

	store := k.getStore(ctx)
	if commission.Commission.IsZero() {
		bz = k.cdc.MustMarshal(&types.ValidatorAccumulatedCommission{})
	} else {
		bz = k.cdc.MustMarshal(&commission)
	}

	store.Set(types.GetValidatorAccumulatedCommissionKey(val), bz)
}

// delete accumulated commission for a validator
func (k Keeper) DeleteValidatorAccumulatedCommission(ctx sdk.Context, val sdk.ValAddress) {
	store := k.getStore(ctx)
	store.Delete(types.GetValidatorAccumulatedCommissionKey(val))
}

// iterate over accumulated commissions
func (k Keeper) IterateValidatorAccumulatedCommissions(ctx sdk.Context, handler func(val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorAccumulatedCommissionPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var commission types.ValidatorAccumulatedCommission
		k.cdc.MustUnmarshal(iter.Value(), &commission)
		addr := types.GetValidatorAccumulatedCommissionAddress(iter.Key())
		if handler(addr, commission) {
			break
		}
	}
}

func (k Keeper) decodeValOutstandingRewards(bz []byte) (types.ValidatorOutstandingRewards, error) {
	var rewards types.ValidatorOutstandingRewards
	if bz == nil {
		return types.ValidatorOutstandingRewards{}, nil
	}
	k.cdc.MustUnmarshal(bz, &rewards)
	return rewards, nil
}

// get validator outstanding rewards
func (k Keeper) GetValidatorOutstandingRewards(ctx sdk.Context, val sdk.ValAddress) (rewards types.ValidatorOutstandingRewards) {
	store := ctx.KVStore(k.storeKey)
	rewards, err := store2.GetAndDecode(store, k.decodeValOutstandingRewards, types.GetValidatorOutstandingRewardsKey(val))
	if err != nil {
		panic(err)
	}
	return
}

// set validator outstanding rewards
func (k Keeper) SetValidatorOutstandingRewards(ctx sdk.Context, val sdk.ValAddress, rewards types.ValidatorOutstandingRewards) {
	store := k.getStore(ctx)
	b := k.cdc.MustMarshal(&rewards)
	store.Set(types.GetValidatorOutstandingRewardsKey(val), b)
}

// delete validator outstanding rewards
func (k Keeper) DeleteValidatorOutstandingRewards(ctx sdk.Context, val sdk.ValAddress) {
	store := k.getStore(ctx)
	store.Delete(types.GetValidatorOutstandingRewardsKey(val))
}

// iterate validator outstanding rewards
func (k Keeper) IterateValidatorOutstandingRewards(ctx sdk.Context, handler func(val sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorOutstandingRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		rewards := types.ValidatorOutstandingRewards{}
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		addr := types.GetValidatorOutstandingRewardsAddress(iter.Key())
		if handler(addr, rewards) {
			break
		}
	}
}

func (k Keeper) decodeSlashEvent(bz []byte) (types.ValidatorSlashEvent, bool) {
	var event types.ValidatorSlashEvent
	if bz == nil {
		return types.ValidatorSlashEvent{}, false
	}
	k.cdc.MustUnmarshal(bz, &event)
	return event, true
}

// get slash event for height
func (k Keeper) GetValidatorSlashEvent(ctx sdk.Context, val sdk.ValAddress, height, period uint64) (event types.ValidatorSlashEvent, found bool) {
	store := ctx.KVStore(k.storeKey)
	event, found = store2.GetAndDecodeWithBool(store, k.decodeSlashEvent, types.GetValidatorSlashEventKey(val, height, period))
	if !found {
		return event, found
	}
	return event, found
}

// set slash event for height
func (k Keeper) SetValidatorSlashEvent(ctx sdk.Context, val sdk.ValAddress, height, period uint64, event types.ValidatorSlashEvent) {
	store := k.getStore(ctx)
	b := k.cdc.MustMarshal(&event)
	store.Set(types.GetValidatorSlashEventKey(val, height, period), b)
}

// iterate over slash events between heights, inclusive
func (k Keeper) IterateValidatorSlashEventsBetween(ctx sdk.Context, val sdk.ValAddress, startingHeight uint64, endingHeight uint64,
	handler func(height uint64, event types.ValidatorSlashEvent) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iter := store.Iterator(
		types.GetValidatorSlashEventKeyPrefix(val, startingHeight),
		types.GetValidatorSlashEventKeyPrefix(val, endingHeight+1),
	)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var event types.ValidatorSlashEvent
		k.cdc.MustUnmarshal(iter.Value(), &event)
		_, height := types.GetValidatorSlashEventAddressHeight(iter.Key())
		if handler(height, event) {
			break
		}
	}
}

// iterate over all slash events
func (k Keeper) IterateValidatorSlashEvents(ctx sdk.Context, handler func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorSlashEventPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var event types.ValidatorSlashEvent
		k.cdc.MustUnmarshal(iter.Value(), &event)
		val, height := types.GetValidatorSlashEventAddressHeight(iter.Key())
		if handler(val, height, event) {
			break
		}
	}
}

// delete slash events for a particular validator
func (k Keeper) DeleteValidatorSlashEvents(ctx sdk.Context, val sdk.ValAddress) {
	store := k.getStore(ctx)
	iter := sdk.KVStorePrefixIterator(store, types.GetValidatorSlashEventPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// delete all slash events
func (k Keeper) DeleteAllValidatorSlashEvents(ctx sdk.Context) {
	store := k.getStore(ctx)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorSlashEventPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
