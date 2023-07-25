package keeper

import (
	"context"
	"cosmossdk.io/math"
	"encoding/json"
	"errors"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
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
func (k Keeper) SetPreviousProposerConsAddr(ctx context.Context, consAddr sdk.ConsAddress) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&gogotypes.BytesValue{Value: consAddr})
	return store.Set(types.ProposerKey, bz)
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
func (k Keeper) SetWinningGrants(ctx context.Context, winningGrants types.WinningGrants) {
	k.Logger(ctx).Info("Setting winning grants", "winning_grants", winningGrants)
	store := runtime.KVStoreAdapter(k.storeKey)
	//b := k.(&winningGrants)
	// marshal the winning grants to JSON
	b, _ := json.Marshal(winningGrants)
	k.Logger(ctx).Info("Setting winning grants", "winning_grants", b)
	//store.Set(types.GetWinningGrantsHeightKey(), b)
	store.Set(types.WinningGrantsKey, b)
}

func (k Keeper) GetWinningGrants(ctx context.Context) (winningGrants types.WinningGrants) {
	k.Logger(ctx).Info("Getting winning grants", "winning_grants")
	store := runtime.KVStoreAdapter(k.storeKey)
	b := store.Get(types.WinningGrantsKey)
	k.Logger(ctx).Info("Getting winning grants b", "winning_grants", b)
	if b == nil {
		return nil
	}
	err := json.Unmarshal(b, &winningGrants)
	if err != nil {
		return nil
	}
	k.Logger(ctx).Info("Getting winning grants", "winning_grants", winningGrants)
	return winningGrants
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

func (k Keeper) GetPreviousProposerReward(ctx context.Context) math.LegacyDec {
	store := runtime.KVStoreAdapter(k.storeKey)
	bz := store.Get(types.ProposerRewardKey)
	if bz == nil {
		panic("previous proposer reward not set")
	}

	addrValue := gogotypes.StringValue{}
	k.cdc.MustUnmarshal(bz, &addrValue)
	value := math.LegacyMustNewDecFromStr(addrValue.GetValue())
	return value
}

// set the proposer public key for this block
func (k Keeper) SetPreviousProposerReward(ctx context.Context, reward math.LegacyDec) {
	store := runtime.KVStoreAdapter(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.StringValue{Value: reward.String()})
	store.Set(types.ProposerRewardKey, bz)
}

func (k Keeper) GetGovernanceContractAddress(ctx context.Context) (address string) {
	store := runtime.KVStoreAdapter(k.storeKey)
	bz := store.Get(types.GovernanceContractAddress)
	addrValue := gogotypes.StringValue{}
	k.cdc.MustUnmarshal(bz, &addrValue)
	return addrValue.GetValue()
}

// SetGovernanceContractAddress sets the governance contract address
func (k Keeper) SetGovernanceContractAddress(ctx context.Context, address string) {
	store := runtime.KVStoreAdapter(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.StringValue{Value: address})
	store.Set(types.GovernanceContractAddress, bz)
}
