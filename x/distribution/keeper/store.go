package keeper

import (
	"context"
	"errors"

	gogotypes "github.com/cosmos/gogoproto/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// GetDelegatorWithdrawAddr get the delegator withdraw address, defaulting to the delegator address
func (k Keeper) GetDelegatorWithdrawAddr(ctx context.Context, delAddr sdk.AccAddress) (sdk.AccAddress, error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetDelegatorWithdrawAddrKey(delAddr))
	if b == nil {
		return delAddr, err
	}
	return b, nil
}

// SetDelegatorWithdrawAddr sets the delegator withdraw address
func (k Keeper) SetDelegatorWithdrawAddr(ctx context.Context, delAddr, withdrawAddr sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetDelegatorWithdrawAddrKey(delAddr), withdrawAddr.Bytes())
}

// DeleteDelegatorWithdrawAddr deletes a delegator withdraw addr
func (k Keeper) DeleteDelegatorWithdrawAddr(ctx context.Context, delAddr, withdrawAddr sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetDelegatorWithdrawAddrKey(delAddr))
}

// IterateDelegatorWithdrawAddrs iterates over delegator withdraw addrs
func (k Keeper) IterateDelegatorWithdrawAddrs(ctx context.Context, handler func(del, addr sdk.AccAddress) (stop bool)) {
	store := k.storeService.OpenKVStore(ctx)
	iter := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.DelegatorWithdrawAddrPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		addr := sdk.AccAddress(iter.Value())
		del := types.GetDelegatorWithdrawInfoAddress(iter.Key())
		if handler(del, addr) {
			break
		}
	}
}

// GetPreviousProposerConsAddr returns the proposer consensus address for the
// current block.
func (k Keeper) GetPreviousProposerConsAddr(ctx context.Context) (sdk.ConsAddress, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ProposerKey)
	if err != nil {
		return nil, err
	}

	if bz == nil {
		return nil, errors.New("previous proposer not set")
	}

	addrValue := gogotypes.BytesValue{}
	err = k.cdc.Unmarshal(bz, &addrValue)
	if err != nil {
		return nil, err
	}

	return addrValue.GetValue(), nil
}

// SetPreviousProposerConsAddr sets the proposer public key for this block
func (k Keeper) SetPreviousProposerConsAddr(ctx context.Context, consAddr sdk.ConsAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&gogotypes.BytesValue{Value: consAddr})
	return store.Set(types.ProposerKey, bz)
}

// GetDelegatorStartingInfo gets the starting info associated with a delegator
func (k Keeper) GetDelegatorStartingInfo(ctx context.Context, val sdk.ValAddress, del sdk.AccAddress) (period types.DelegatorStartingInfo, err error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetDelegatorStartingInfoKey(val, del))
	if err != nil {
		return
	}

	err = k.cdc.Unmarshal(b, &period)
	return period, err
}

// SetDelegatorStartingInfo sets the starting info associated with a delegator
func (k Keeper) SetDelegatorStartingInfo(ctx context.Context, val sdk.ValAddress, del sdk.AccAddress, period types.DelegatorStartingInfo) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&period)
	if err != nil {
		return err
	}

	return store.Set(types.GetDelegatorStartingInfoKey(val, del), b)
}

// HasDelegatorStartingInfo checks existence of the starting info associated with a delegator
func (k Keeper) HasDelegatorStartingInfo(ctx context.Context, val sdk.ValAddress, del sdk.AccAddress) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Has(types.GetDelegatorStartingInfoKey(val, del))
}

// DeleteDelegatorStartingInfo deletes the starting info associated with a delegator
func (k Keeper) DeleteDelegatorStartingInfo(ctx context.Context, val sdk.ValAddress, del sdk.AccAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetDelegatorStartingInfoKey(val, del))
}

// IterateDelegatorStartingInfos iterates over delegator starting infos
func (k Keeper) IterateDelegatorStartingInfos(ctx context.Context, handler func(val sdk.ValAddress, del sdk.AccAddress, info types.DelegatorStartingInfo) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.DelegatorStartingInfoPrefix)
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

// GetValidatorHistoricalRewards gets historical rewards for a particular period
func (k Keeper) GetValidatorHistoricalRewards(ctx context.Context, val sdk.ValAddress, period uint64) (rewards types.ValidatorHistoricalRewards, err error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetValidatorHistoricalRewardsKey(val, period))
	if err != nil {
		return
	}

	err = k.cdc.Unmarshal(b, &rewards)
	return
}

// SetValidatorHistoricalRewards sets historical rewards for a particular period
func (k Keeper) SetValidatorHistoricalRewards(ctx context.Context, val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&rewards)
	if err != nil {
		return err
	}

	return store.Set(types.GetValidatorHistoricalRewardsKey(val, period), b)
}

// IterateValidatorHistoricalRewards iterates over historical rewards
func (k Keeper) IterateValidatorHistoricalRewards(ctx context.Context, handler func(val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
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

// DeleteValidatorHistoricalReward deletes a historical reward
func (k Keeper) DeleteValidatorHistoricalReward(ctx context.Context, val sdk.ValAddress, period uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetValidatorHistoricalRewardsKey(val, period))
}

// DeleteValidatorHistoricalRewards deletes historical rewards for a validator
func (k Keeper) DeleteValidatorHistoricalRewards(ctx context.Context, val sdk.ValAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.GetValidatorHistoricalRewardsPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// DeleteAllValidatorHistoricalRewards deletes all historical rewards
func (k Keeper) DeleteAllValidatorHistoricalRewards(ctx context.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// GetValidatorHistoricalReferenceCount gets a historical reference count (used for testcases)
func (k Keeper) GetValidatorHistoricalReferenceCount(ctx context.Context) (count uint64) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var rewards types.ValidatorHistoricalRewards
		k.cdc.MustUnmarshal(iter.Value(), &rewards)
		count += uint64(rewards.ReferenceCount)
	}
	return
}

// GetValidatorCurrentRewards gets current rewards for a validator
func (k Keeper) GetValidatorCurrentRewards(ctx context.Context, val sdk.ValAddress) (rewards types.ValidatorCurrentRewards, err error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetValidatorCurrentRewardsKey(val))
	if err != nil {
		return
	}

	err = k.cdc.Unmarshal(b, &rewards)
	return
}

// SetValidatorCurrentRewards sets current rewards for a validator
func (k Keeper) SetValidatorCurrentRewards(ctx context.Context, val sdk.ValAddress, rewards types.ValidatorCurrentRewards) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&rewards)
	if err != nil {
		return err
	}

	return store.Set(types.GetValidatorCurrentRewardsKey(val), b)
}

// DeleteValidatorCurrentRewards deletes current rewards for a validator
func (k Keeper) DeleteValidatorCurrentRewards(ctx context.Context, val sdk.ValAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetValidatorCurrentRewardsKey(val))
}

// IterateValidatorCurrentRewards iterates over current rewards
func (k Keeper) IterateValidatorCurrentRewards(ctx context.Context, handler func(val sdk.ValAddress, rewards types.ValidatorCurrentRewards) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorCurrentRewardsPrefix)
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

// GetValidatorAccumulatedCommission gets accumulated commission for a validator
func (k Keeper) GetValidatorAccumulatedCommission(ctx context.Context, val sdk.ValAddress) (commission types.ValidatorAccumulatedCommission, err error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetValidatorAccumulatedCommissionKey(val))
	if err != nil {
		return types.ValidatorAccumulatedCommission{}, err
	}

	if b == nil {
		return types.ValidatorAccumulatedCommission{}, nil
	}

	err = k.cdc.Unmarshal(b, &commission)
	if err != nil {
		return types.ValidatorAccumulatedCommission{}, err
	}
	return
}

// SetValidatorAccumulatedCommission sets accumulated commission for a validator
func (k Keeper) SetValidatorAccumulatedCommission(ctx context.Context, val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) error {
	var (
		bz  []byte
		err error
	)

	store := k.storeService.OpenKVStore(ctx)
	if commission.Commission.IsZero() {
		bz, err = k.cdc.Marshal(&types.ValidatorAccumulatedCommission{})
	} else {
		bz, err = k.cdc.Marshal(&commission)
	}

	if err != nil {
		return err
	}

	return store.Set(types.GetValidatorAccumulatedCommissionKey(val), bz)
}

// DeleteValidatorAccumulatedCommission deletes accumulated commission for a validator
func (k Keeper) DeleteValidatorAccumulatedCommission(ctx context.Context, val sdk.ValAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetValidatorAccumulatedCommissionKey(val))
}

// IterateValidatorAccumulatedCommissions iterates over accumulated commissions
func (k Keeper) IterateValidatorAccumulatedCommissions(ctx context.Context, handler func(val sdk.ValAddress, commission types.ValidatorAccumulatedCommission) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorAccumulatedCommissionPrefix)
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

// GetValidatorOutstandingRewards gets validator outstanding rewards
func (k Keeper) GetValidatorOutstandingRewards(ctx context.Context, val sdk.ValAddress) (rewards types.ValidatorOutstandingRewards, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetValidatorOutstandingRewardsKey(val))
	if err != nil {
		return
	}
	err = k.cdc.Unmarshal(bz, &rewards)
	return
}

// SetValidatorOutstandingRewards sets validator outstanding rewards
func (k Keeper) SetValidatorOutstandingRewards(ctx context.Context, val sdk.ValAddress, rewards types.ValidatorOutstandingRewards) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&rewards)
	if err != nil {
		return err
	}
	return store.Set(types.GetValidatorOutstandingRewardsKey(val), b)
}

// DeleteValidatorOutstandingRewards deletes validator outstanding rewards
func (k Keeper) DeleteValidatorOutstandingRewards(ctx context.Context, val sdk.ValAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetValidatorOutstandingRewardsKey(val))
}

// IterateValidatorOutstandingRewards iterates validator outstanding rewards
func (k Keeper) IterateValidatorOutstandingRewards(ctx context.Context, handler func(val sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorOutstandingRewardsPrefix)
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

// GetValidatorSlashEvent gets slash event for height
func (k Keeper) GetValidatorSlashEvent(ctx context.Context, val sdk.ValAddress, height, period uint64) (event types.ValidatorSlashEvent, found bool, err error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetValidatorSlashEventKey(val, height, period))
	if err != nil {
		return types.ValidatorSlashEvent{}, false, err
	}

	if b == nil {
		return types.ValidatorSlashEvent{}, false, nil
	}

	err = k.cdc.Unmarshal(b, &event)
	if err != nil {
		return types.ValidatorSlashEvent{}, false, err
	}

	return event, true, nil
}

// SetValidatorSlashEvent sets slash event for height
func (k Keeper) SetValidatorSlashEvent(ctx context.Context, val sdk.ValAddress, height, period uint64, event types.ValidatorSlashEvent) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&event)
	if err != nil {
		return err
	}

	return store.Set(types.GetValidatorSlashEventKey(val, height, period), b)
}

// IterateValidatorSlashEventsBetween iterates over slash events between heights, inclusive
func (k Keeper) IterateValidatorSlashEventsBetween(ctx context.Context, val sdk.ValAddress, startingHeight, endingHeight uint64,
	handler func(height uint64, event types.ValidatorSlashEvent) (stop bool),
) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
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

// IterateValidatorSlashEvents iterates over all slash events
func (k Keeper) IterateValidatorSlashEvents(ctx context.Context, handler func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorSlashEventPrefix)
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

// DeleteValidatorSlashEvents deletes slash events for a particular validator
func (k Keeper) DeleteValidatorSlashEvents(ctx context.Context, val sdk.ValAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.GetValidatorSlashEventPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// DeleteAllValidatorSlashEvents deletes all slash events
func (k Keeper) DeleteAllValidatorSlashEvents(ctx context.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorSlashEventPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
