package keeper

import (
	"context"
	"errors"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MaxConsKeyRotations is the maximum number of pending consensus key rotations
// a validator may have inside the unbonding window.
const MaxConsKeyRotations = 1

// HasConsKeyRotationInUnbondingWindow returns whether the validator has
// performed a consensus key rotation inside current the unbonding window.
func (k Keeper) HasConsKeyRotationInUnbondingWindow(ctx context.Context, valAddr sdk.ValAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetValidatorConsKeyRotationKey(valAddr))
}

// IsConsAddrLockedByRotation returns whether the given consensus address is
// locked by a key rotation, either because some validator previously rotated
// away from it (and is still inside the unbonding window) or because some
// validator has enqueued a pending rotation targeting it.
func (k Keeper) IsConsAddrLockedByRotation(ctx context.Context, consAddr sdk.ConsAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetRotationLockedConsAddrIndexKey(consAddr))
}

// HasConsKeyRotationQueueEntry returns whether the maturity queue holds an
// entry at the given maturity for the given validator.
func (k Keeper) HasConsKeyRotationQueueEntry(ctx context.Context, maturity time.Time, valAddr sdk.ValAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetConsKeyRotationQueueKey(maturity, valAddr))
}

// SetConsKeyRotation writes to indexes that track a pending consensus key
// rotation. The new pubkey is written to the unapplied queue so the end
// blocker can perform the rotation in this block.
func (k Keeper) SetConsKeyRotation(ctx context.Context, valAddr sdk.ValAddress, oldPubKey, newPubKey cryptotypes.PubKey, fee sdk.Coin) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	unbondingTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return err
	}
	maturity := sdkCtx.BlockHeader().Time.Add(unbondingTime)

	oldConsAddr := sdk.ConsAddress(oldPubKey.Address())
	newConsAddr := sdk.ConsAddress(newPubKey.Address())

	store := k.storeService.OpenKVStore(ctx)

	// add to queue keyed by time so that we can iterate rotations happening by
	// time and quickly remove ones that have matured (fallen out of the
	// current unbonding period).
	if err := store.Set(types.GetConsKeyRotationQueueKey(maturity, valAddr), oldConsAddr); err != nil {
		return err
	}

	if err := store.Set(types.GetValidatorConsKeyRotationKey(valAddr), []byte{}); err != nil {
		return err
	}

	// lock both the old and new cons addrs so that no validator can rotate
	// to either while the rotation is pending or within the unbonding
	// window. The new addr entry is cleared when the rotation is applied
	// in the end blocker (after which it is the validator's live cons
	// addr). The old addr entry is cleared when the rotation matures.
	if err := store.Set(types.GetRotationLockedConsAddrIndexKey(oldConsAddr), valAddr); err != nil {
		return err
	}
	if err := store.Set(types.GetRotationLockedConsAddrIndexKey(newConsAddr), valAddr); err != nil {
		return err
	}

	newPubKeyBz, err := k.cdc.MarshalInterface(newPubKey)
	if err != nil {
		return err
	}
	return store.Set(types.GetUnappliedConsKeyRotationKey(valAddr), newPubKeyBz)
}

// ApplyPendingConsKeyRotations applies every rotation queued by the msg server
// and returns the validator updates needed to retire each old key at zero
// power and instate each new key at the validator's current power.
func (k Keeper) ApplyPendingConsKeyRotations(ctx context.Context, powerReduction math.Int) ([]abci.ValidatorUpdate, error) {
	var totalUpdates abci.ValidatorUpdates

	store := k.storeService.OpenKVStore(ctx)
	err := k.IterateUnappliedConsKeyRotations(ctx, func(valAddr sdk.ValAddress, newPubKey cryptotypes.PubKey) error {
		validator, err := k.GetValidator(ctx, valAddr)
		if err != nil {
			return err
		}

		// handles updating state with the validators new consensus key and
		// creating abci updates to pass to comet
		updates, err := k.ApplyConsKeyRotation(ctx, validator, newPubKey, powerReduction)
		if err != nil {
			return err
		}
		totalUpdates = append(totalUpdates, updates...)

		// the new cons addr is now the validator's live cons addr; further
		// rotations targeting it are blocked by the by cons addr lookup,
		// so release its rotation lock entry. The old cons addr entry
		// stays until the rotation matures.
		if err := store.Delete(types.GetRotationLockedConsAddrIndexKey(sdk.ConsAddress(newPubKey.Address()))); err != nil {
			return err
		}

		return store.Delete(types.GetUnappliedConsKeyRotationKey(valAddr))
	})
	if err != nil {
		return nil, err
	}

	return totalUpdates, nil
}

// ApplyConsKeyRotation switches the validator's consensus pubkey to newPubKey
// in x/staking state. The validator record is updated, the old by cons address
// index entry is deleted, and the new by cons address index entry is written.
func (k Keeper) ApplyConsKeyRotation(ctx context.Context, validator types.Validator, newPubKey cryptotypes.PubKey, powerReduction math.Int) (abci.ValidatorUpdates, error) {
	// we will have two validator updates for every consensus key rotation
	// since a key rotation to comet looks like a validator becoming 0 power
	// (the old cons addr) and a new validator coming online with the new cons
	// addr that has the same power as the old validator.
	updates := make([]abci.ValidatorUpdate, 2)

	// create a validator update that will mark its current cons addr as 0
	// power
	updates[0] = validator.ABCIValidatorUpdateZero()

	// capture the old cons addr before the in-memory swap below, so that we
	// can delete the old by cons addr index entry further down
	oldConsAddr, err := validator.GetConsAddr()
	if err != nil {
		return nil, err
	}

	// update the validator in memory to use the new cons addr
	newAny, err := codectypes.NewAnyWithValue(newPubKey)
	if err != nil {
		return nil, err
	}
	validator.ConsensusPubkey = newAny

	// create a validator update that will mark its new cons addr with the same
	// power as its previous cons addr
	updates[1] = validator.ABCIValidatorUpdate(powerReduction)

	// set the validator in the store (keyed by operator address, which didnt
	// change) to the updated validator with the new cons addr
	if err := k.SetValidator(ctx, validator); err != nil {
		return nil, err
	}

	// remove the store entry for the previous cons addr pointing to the
	// validator
	store := k.storeService.OpenKVStore(ctx)
	if err := store.Delete(types.GetValidatorByConsAddrKey(oldConsAddr)); err != nil {
		return nil, err
	}

	// create a new store entry for the new cons addr pointing to the validator
	if err := k.SetValidatorByConsAddr(ctx, validator); err != nil {
		return nil, err
	}

	return updates, nil
}

// PruneMaturedConsKeyRotations removes every rotation whose unbonding window
// has elapsed at the current block time. It deletes the entries from the
// maturity queue, the per validator pending index, and the rotated consensus
// address index.
func (k Keeper) PruneMaturedConsKeyRotations(ctx context.Context) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockHeader().Time

	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(
		types.ConsKeyRotationQueueKey,
		storetypes.PrefixEndBytes(types.GetConsKeyRotationQueueTimePrefix(blockTime)),
	)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	for ; iterator.Valid(); iterator.Next() {
		// TODO: migrate ValidatorSigningInfo from oldConsAddr to newConsAddr
		_, valAddr, err := types.ParseConsKeyRotationQueueKey(iterator.Key())
		if err != nil {
			return err
		}
		oldConsAddr := sdk.ConsAddress(iterator.Value())

		if err := store.Delete(iterator.Key()); err != nil {
			return err
		}
		if err := store.Delete(types.GetValidatorConsKeyRotationKey(valAddr)); err != nil {
			return err
		}
		if err := store.Delete(types.GetRotationLockedConsAddrIndexKey(oldConsAddr)); err != nil {
			return err
		}
	}

	return nil
}

// IterateUnappliedConsKeyRotations walks every rotation queued by the msg
// server that the end blocker has not yet applied, in valAddr sorted order.
func (k Keeper) IterateUnappliedConsKeyRotations(
	ctx context.Context,
	cb func(valAddr sdk.ValAddress, newPubKey cryptotypes.PubKey) error,
) (err error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.UnappliedConsKeyRotationKey, storetypes.PrefixEndBytes(types.UnappliedConsKeyRotationKey))
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		valAddr := sdk.ValAddress(key[len(types.UnappliedConsKeyRotationKey)+1:])

		var newPubKey cryptotypes.PubKey
		if err := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); err != nil {
			return err
		}

		if err := cb(valAddr, newPubKey); err != nil {
			return err
		}
	}
	return nil
}
