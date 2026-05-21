package keeper

import (
	"context"
	"errors"

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

// HasPendingConsKeyRotation returns whether the validator has a pending
// consensus key rotation inside the unbonding window.
func (k Keeper) HasPendingConsKeyRotation(ctx context.Context, valAddr sdk.ValAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetValidatorConsKeyRotationKey(valAddr))
}

// HasRotatedConsAddr returns whether the given consensus address was previously
// rotated away from and is still inside its unbonding window.
func (k Keeper) HasRotatedConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetRotatedConsAddrIndexKey(consAddr))
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

	if err := store.Set(types.GetRotatedConsAddrIndexKey(oldConsAddr), valAddr); err != nil {
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
	var (
		// used to defer the removal of UnappliedConsensusKeyRotationKeys until
		// after iteration
		rotatedValidators []sdk.ValAddress

		totalUpdates abci.ValidatorUpdates
	)

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

		// defer removal until after iteration
		rotatedValidators = append(rotatedValidators, valAddr)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// perform removal of pending rotation for each validator that rotated
	for _, rotatedValidator := range rotatedValidators {
		if err := store.Delete(types.GetUnappliedConsKeyRotationKey(rotatedValidator)); err != nil {
			return nil, err
		}
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

	store := k.storeService.OpenKVStore(ctx)
	oldConsAddr, err := validator.GetConsAddr()
	if err != nil {
		return nil, err
	}

	// remove the store entry for the previous cons addr pointing to the
	// validator
	if err := store.Delete(types.GetValidatorByConsAddrKey(oldConsAddr)); err != nil {
		return nil, err
	}

	// create a new store entry for the new cons addr pointing to the validator
	if err := k.SetValidatorByConsAddr(ctx, validator); err != nil {
		return nil, err
	}

	return updates, nil
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
