package keeper

import (
	"context"
	"errors"
	"math"

	"cosmossdk.io/collections"

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

// historical reference count (used for testcases)
func (k Keeper) GetValidatorHistoricalReferenceCount(ctx context.Context) (count uint64) {
	iter, err := k.ValidatorHistoricalRewards.Iterate(
		ctx, nil,
	)

	if errors.Is(err, collections.ErrInvalidIterator) {
		return
	}

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		rewards, err := iter.Value()
		if err != nil {
			panic(err)
		}
		count += uint64(rewards.ReferenceCount)
	}
	return
}

// iterate over slash events between heights, inclusive
func (k Keeper) IterateValidatorSlashEventsBetween(ctx context.Context, val sdk.ValAddress, startingHeight, endingHeight uint64,
	handler func(height uint64, event types.ValidatorSlashEvent) (stop bool),
) {
	rng := new(collections.Range[collections.Triple[sdk.ValAddress, uint64, uint64]]).
		StartInclusive(collections.Join3(val, startingHeight, uint64(0))).
		EndExclusive(collections.Join3(val, endingHeight+1, uint64(math.MaxUint64)))

	iter, err := k.ValidatorSlashEvents.Iterate(
		ctx,
		rng,
	)

	if errors.Is(err, collections.ErrInvalidIterator) {
		return
	}

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		event, err := iter.Value()
		if err != nil {
			panic(err)
		}
		k, err := iter.Key()
		if err != nil {
			panic(err)
		}

		height := k.K2()
		if handler(height, event) {
			break
		}
	}
}

// iterate over all slash events
func (k Keeper) IterateValidatorSlashEvents(ctx context.Context, handler func(val sdk.ValAddress, height uint64, event types.ValidatorSlashEvent) (stop bool)) {
	iter, err := k.ValidatorSlashEvents.Iterate(ctx, nil)
	if errors.Is(err, collections.ErrInvalidIterator) {
		return
	}

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		event, err := iter.Value()
		if err != nil {
			panic(err)
		}

		k, err := iter.Key()
		if err != nil {
			panic(err)
		}

		if handler(k.K1(), k.K2(), event) {
			break
		}
	}
}

// delete slash events for a particular validator
func (k Keeper) DeleteValidatorSlashEvents(ctx context.Context, val sdk.ValAddress) {
	err := k.ValidatorSlashEvents.Clear(ctx, collections.NewPrefixedTripleRange[sdk.ValAddress, uint64, uint64](val))
	if err != nil {
		panic(err)
	}
}

// delete all slash events
func (k Keeper) DeleteAllValidatorSlashEvents(ctx context.Context) {
	err := k.ValidatorSlashEvents.Clear(ctx, nil)
	if err != nil {
		panic(err)
	}
}
