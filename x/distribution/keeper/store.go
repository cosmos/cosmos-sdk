package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	gogotypes "github.com/cosmos/gogoproto/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// get the delegator withdraw address, defaulting to the delegator address
func (k Keeper) GetDelegatorWithdrawAddr(ctx context.Context, delAddr sdk.AccAddress) (sdk.AccAddress, error) {
	addr, err := k.DelegatorsWithdrawAddress.Get(ctx, delAddr)
	if err != nil && errors.Is(err, collections.ErrNotFound) {
		return delAddr, nil
	}
	return addr, err
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

// set the proposer public key for this block
func (k Keeper) SetPreviousProposerConsAddr(ctx context.Context, consAddr sdk.ConsAddress) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&gogotypes.BytesValue{Value: consAddr})
	store.Set(types.ProposerKey, bz)
}

// get historical rewards for a particular period
func (k Keeper) GetValidatorHistoricalRewards(ctx context.Context, val sdk.ValAddress, period uint64) (rewards types.ValidatorHistoricalRewards, err error) {
	store := k.storeService.OpenKVStore(ctx)
	b, err := store.Get(types.GetValidatorHistoricalRewardsKey(val, period))
	if err != nil {
		return
	}

	err = k.cdc.Unmarshal(b, &rewards)
	return
}

// set historical rewards for a particular period
func (k Keeper) SetValidatorHistoricalRewards(ctx context.Context, val sdk.ValAddress, period uint64, rewards types.ValidatorHistoricalRewards) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&rewards)
	if err != nil {
		return err
	}

	return store.Set(types.GetValidatorHistoricalRewardsKey(val, period), b)
}

// iterate over historical rewards
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

// delete a historical reward
func (k Keeper) DeleteValidatorHistoricalReward(ctx context.Context, val sdk.ValAddress, period uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.GetValidatorHistoricalRewardsKey(val, period))
}

// delete historical rewards for a validator
func (k Keeper) DeleteValidatorHistoricalRewards(ctx context.Context, val sdk.ValAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.GetValidatorHistoricalRewardsPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// delete all historical rewards
func (k Keeper) DeleteAllValidatorHistoricalRewards(ctx context.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorHistoricalRewardsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// historical reference count (used for testcases)
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

// get validator outstanding rewards
func (k Keeper) GetValidatorOutstandingRewards(ctx context.Context, val sdk.ValAddress) (rewards types.ValidatorOutstandingRewards, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetValidatorOutstandingRewardsKey(val))
	if err != nil {
		return
	}
	err = k.cdc.Unmarshal(bz, &rewards)
	return
}

// get slash event for height
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

// set slash event for height
func (k Keeper) SetValidatorSlashEvent(ctx context.Context, val sdk.ValAddress, height, period uint64, event types.ValidatorSlashEvent) error {
	store := k.storeService.OpenKVStore(ctx)
	b, err := k.cdc.Marshal(&event)
	if err != nil {
		return err
	}

	return store.Set(types.GetValidatorSlashEventKey(val, height, period), b)
}

// iterate over slash events between heights, inclusive
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

// iterate over all slash events
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

// delete slash events for a particular validator
func (k Keeper) DeleteValidatorSlashEvents(ctx context.Context, val sdk.ValAddress) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.GetValidatorSlashEventPrefix(val))
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// delete all slash events
func (k Keeper) DeleteAllValidatorSlashEvents(ctx context.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iter := storetypes.KVStorePrefixIterator(store, types.ValidatorSlashEventPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}
