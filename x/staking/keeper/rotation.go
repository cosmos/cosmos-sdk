package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ImportConsKeyRotations restores consensus key rotation indexes and queues
// from genesis.
func (k Keeper) ImportConsKeyRotations(ctx context.Context, histories []types.ConsensusKeyRotationHistory, pending []types.PendingConsensusKeyRotation) error {
	store := k.storeService.OpenKVStore(ctx)

	for _, history := range histories {
		valAddr, err := k.validatorAddressCodec.StringToBytes(history.ValidatorAddress)
		if err != nil {
			return err
		}
		oldConsAddr, err := k.consensusAddressCodec.StringToBytes(history.OldConsensusAddress)
		if err != nil {
			return err
		}

		if err := store.Set(types.GetConsKeyRotationQueueKey(history.MaturityTime, valAddr), oldConsAddr); err != nil {
			return err
		}
		if err := store.Set(types.GetValidatorConsKeyRotationKey(valAddr), []byte{}); err != nil {
			return err
		}
		if err := store.Set(types.GetRotationLockedConsAddrIndexKey(oldConsAddr), valAddr); err != nil {
			return err
		}
	}

	for _, rotation := range pending {
		valAddr, err := k.validatorAddressCodec.StringToBytes(rotation.ValidatorAddress)
		if err != nil {
			return err
		}

		var newPubKey cryptotypes.PubKey
		if err := k.cdc.UnpackAny(rotation.NewPubkey, &newPubKey); err != nil {
			return err
		}
		newPubKeyBz, err := k.cdc.MarshalInterface(newPubKey)
		if err != nil {
			return err
		}

		if err := store.Set(types.GetRotationLockedConsAddrIndexKey(sdk.ConsAddress(newPubKey.Address())), valAddr); err != nil {
			return err
		}
		if err := store.Set(types.GetConsKeyRotationApplyQueueKey(rotation.ApplyHeight, valAddr), newPubKeyBz); err != nil {
			return err
		}
	}

	return nil
}

// ExportConsKeyRotationHistory returns consensus key rotation history records
// that are still inside the unbonding window.
func (k Keeper) ExportConsKeyRotationHistory(ctx context.Context) (histories []types.ConsensusKeyRotationHistory, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ConsKeyRotationQueueKey, storetypes.PrefixEndBytes(types.ConsKeyRotationQueueKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	for ; iterator.Valid(); iterator.Next() {
		maturity, valAddr, err := types.ParseConsKeyRotationQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		valAddrStr, err := k.validatorAddressCodec.BytesToString(valAddr)
		if err != nil {
			return nil, err
		}
		oldConsAddrStr, err := k.consensusAddressCodec.BytesToString(iterator.Value())
		if err != nil {
			return nil, err
		}

		histories = append(histories, types.ConsensusKeyRotationHistory{
			ValidatorAddress:    valAddrStr,
			OldConsensusAddress: oldConsAddrStr,
			MaturityTime:        maturity,
		})
	}

	return histories, nil
}

// ExportPendingConsKeyRotations returns consensus key rotations whose deferred
// SDK side state update has not been applied yet. Rotations that are not
// active at exportedInitialHeight are pushed forward so an imported chain can
// reemit their validator update before applying the SDK side key swap.
func (k Keeper) ExportPendingConsKeyRotations(ctx context.Context, exportedInitialHeight int64) (rotations []types.PendingConsensusKeyRotation, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ConsKeyRotationApplyQueueKey, storetypes.PrefixEndBytes(types.ConsKeyRotationApplyQueueKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	for ; iterator.Valid(); iterator.Next() {
		applyHeight, valAddr, err := types.ParseConsKeyRotationApplyQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		valAddrStr, err := k.validatorAddressCodec.BytesToString(valAddr)
		if err != nil {
			return nil, err
		}
		var newPubKey cryptotypes.PubKey
		if err := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); err != nil {
			return nil, err
		}
		newPubKeyAny, err := codectypes.NewAnyWithValue(newPubKey)
		if err != nil {
			return nil, err
		}

		if applyHeight > exportedInitialHeight {
			// the old chain has already emitted the abci validator update, but
			// that pending comet side transition is not part of app genesis.
			// if we import the genesis without modifying the apply height, we
			// will comet will not know about this key rotation and we will
			// update sdk state without updating comet. thus, we push the apply
			// height forward so hat the abci updates can be reemitted and
			// comet properly updated
			applyHeight = exportedInitialHeight + types.ConsensusUpdateDelay
		}

		rotations = append(rotations, types.PendingConsensusKeyRotation{
			ValidatorAddress: valAddrStr,
			NewPubkey:        newPubKeyAny,
			ApplyHeight:      applyHeight,
		})
	}

	return rotations, nil
}

// PrepareConsKeyRotationsForZeroHeightExport rewrites pending rotation apply
// queue entries so a zero-height restarted chain can emit the Comet validator
// update from its first block before the SDK-side key swap is applied.
func (k Keeper) PrepareConsKeyRotationsForZeroHeightExport(ctx context.Context) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ConsKeyRotationApplyQueueKey, storetypes.PrefixEndBytes(types.ConsKeyRotationApplyQueueKey))
	if err != nil {
		return err
	}

	type pendingRotation struct {
		oldKey  []byte
		valAddr sdk.ValAddress
		value   []byte
	}
	var rotations []pendingRotation

	// collect apply queue entries first since we need to rewrite them, this
	// requires a delete
	for ; iterator.Valid(); iterator.Next() {
		key := append([]byte(nil), iterator.Key()...)
		_, valAddr, err := types.ParseConsKeyRotationApplyQueueKey(key)
		if err != nil {
			return errors.Join(err, iterator.Close())
		}
		rotations = append(rotations, pendingRotation{
			oldKey:  key,
			valAddr: append(sdk.ValAddress(nil), valAddr...),
			value:   append([]byte(nil), iterator.Value()...),
		})
	}
	if err := iterator.Close(); err != nil {
		return err
	}

	// rewrite apply height to in initial height + update delay
	applyHeight := int64(1) + types.ConsensusUpdateDelay
	for _, rotation := range rotations {
		if err := store.Delete(rotation.oldKey); err != nil {
			return err
		}
		if err := store.Set(types.GetConsKeyRotationApplyQueueKey(applyHeight, rotation.valAddr), rotation.value); err != nil {
			return err
		}
	}

	return nil
}

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

// HasConsKeyRotationApplyQueueEntry returns whether the apply queue holds an
// entry at the given apply height for the given validator.
func (k Keeper) HasConsKeyRotationApplyQueueEntry(ctx context.Context, applyHeight int64, valAddr sdk.ValAddress) (bool, error) {
	return k.storeService.OpenKVStore(ctx).Has(types.GetConsKeyRotationApplyQueueKey(applyHeight, valAddr))
}

// SetConsKeyRotation writes the indexes that track a pending consensus key
// rotation.
func (k Keeper) SetConsKeyRotation(ctx context.Context, valAddr sdk.ValAddress, oldPubKey, newPubKey cryptotypes.PubKey) error {
	maturesAt, err := k.rotationMaturityTime(ctx)
	if err != nil {
		return err
	}

	oldConsAddr := sdk.ConsAddress(oldPubKey.Address())
	newConsAddr := sdk.ConsAddress(newPubKey.Address())

	store := k.storeService.OpenKVStore(ctx)

	// add to queue keyed by time so that we can iterate rotations happening by
	// time and quickly remove ones that have matured (fallen out of the
	// current unbonding period).
	if err := store.Set(types.GetConsKeyRotationQueueKey(maturesAt, valAddr), oldConsAddr); err != nil {
		return err
	}

	// mark this validator has having rotated (used to block future rotation
	// for this validator within the current unbonding period).
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

	// add the rotation to the apply queue, so the end blocker can:
	// 1. iterate the queue and emit validator power updates for rotations that
	//    are behind their apply height.
	// 2. iterate the queue and update staking state for validators who should
	//    have their rotations applied to sdk state (at apply height).
	applyHeight := rotationApplyHeight(ctx)
	return store.Set(types.GetConsKeyRotationApplyQueueKey(applyHeight, valAddr), newPubKeyBz)
}

// ProcessConsKeyRotations performs the two passes over the height-keyed apply
// queue called once per EndBlock from ApplyAndReturnValidatorSetUpdates:
//
//   - Drain (mature): for every entry with applyHeight <= currentHeight, apply
//     the SDK-side state swap (validator.ConsensusPubkey and byConsAddr index),
//     delete the queue entry, and release the new-addr rotation lock. The
//     old-addr lock persists until unbonding-window pruning.
//   - Emit (new this block): for every entry with applyHeight ==
//     currentHeight + ConsensusUpdateDelay, build and append the
//     (old@0, new@power) ValidatorUpdate pair. The entry is not deleted here;
//     it remains until its applyHeight matures.
//
// The drain pass runs first so that subsequent transition emits in the same
// EndBlock read the post-swap ConsensusPubkey.
func (k Keeper) ProcessConsKeyRotations(ctx context.Context, powerReduction math.Int) ([]abci.ValidatorUpdate, error) {
	if err := k.ApplyConsKeyRotations(ctx); err != nil {
		return nil, err
	}
	return k.ConsKeyRotationUpdates(ctx, powerReduction)
}

// ApplyConsKeyRotations iterates every apply queue entry whose applyHeight has
// been reached, performs the state swap, and clears the queue entry plus the
// new addr rotation lock.
func (k Keeper) ApplyConsKeyRotations(ctx context.Context) (err error) {
	store := k.storeService.OpenKVStore(ctx)

	// iterate (-, currentHeight+1) which covers every applyHeight <= currentHeight
	currentHeight := sdk.UnwrapSDKContext(ctx).BlockHeight()
	iterator, err := store.Iterator(
		types.ConsKeyRotationApplyQueueKey,
		storetypes.PrefixEndBytes(types.GetConsKeyRotationApplyQueueHeightPrefix(currentHeight)),
	)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	type matured struct {
		key       []byte
		valAddr   sdk.ValAddress
		newPubKey cryptotypes.PubKey
	}
	// collect first; SDK store iterators do not promise safety when keys are
	// deleted mid-iteration.
	var entries []matured
	for ; iterator.Valid(); iterator.Next() {
		keyCopy := append([]byte(nil), iterator.Key()...)
		_, valAddr, err := types.ParseConsKeyRotationApplyQueueKey(keyCopy)
		if err != nil {
			return err
		}
		var newPubKey cryptotypes.PubKey
		if uerr := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); uerr != nil {
			return uerr
		}
		entries = append(entries, matured{key: keyCopy, valAddr: valAddr, newPubKey: newPubKey})
	}

	for _, e := range entries {
		if aerr := k.ApplyConsKeyRotation(ctx, e.valAddr, e.newPubKey); aerr != nil {
			return aerr
		}

		// the new cons addr is now the validator's live cons addr. further
		// rotations targeting it are blocked by the by cons addr lookup, so
		// release its rotation lock entry. The old cons addr entry stays
		// until the rotation matures.
		if derr := store.Delete(types.GetRotationLockedConsAddrIndexKey(sdk.ConsAddress(e.newPubKey.Address()))); derr != nil {
			return derr
		}

		// delete the entry from the apply queue
		if derr := store.Delete(e.key); derr != nil {
			return derr
		}
	}
	return nil
}

// ApplyConsKeyRotation swaps the validator's ConsensusPubkey to newPubKey
// and updates the byConsAddr index. Returns nil silently if the validator no
// longer exists.
func (k Keeper) ApplyConsKeyRotation(ctx context.Context, valAddr sdk.ValAddress, newPubKey cryptotypes.PubKey) error {
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		// this could happen if the validator is removed (unbonds) between when
		// the rotation was enqueued, and now when we are actually applying it
		// (since there is a 2 block delay between enqueue and apply).
		if errors.Is(err, types.ErrNoValidatorFound) {
			return nil
		}
		return err
	}

	oldConsAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}

	newAny, err := codectypes.NewAnyWithValue(newPubKey)
	if err != nil {
		return err
	}
	validator.ConsensusPubkey = newAny

	if err := k.SetValidator(ctx, validator); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	if err := store.Delete(types.GetValidatorByConsAddrKey(oldConsAddr)); err != nil {
		return err
	}
	return k.SetValidatorByConsAddr(ctx, validator)
}

// ConsKeyRotationUpdates returns power updates for each validator
// rotating their consensus keys.
func (k Keeper) ConsKeyRotationUpdates(ctx context.Context, powerReduction math.Int) (updates []abci.ValidatorUpdate, err error) {
	store := k.storeService.OpenKVStore(ctx)

	// iterate all entries in the apply queue that are equal to applyHeight
	applyHeight := rotationApplyHeight(ctx)
	prefix := types.GetConsKeyRotationApplyQueueHeightPrefix(applyHeight)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	for ; iterator.Valid(); iterator.Next() {
		_, valAddr, err := types.ParseConsKeyRotationApplyQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		validator, err := k.GetValidator(ctx, valAddr)
		if err != nil {
			// the validator may have been removed earlier in the same block
			// (e.g., evidence-driven tombstoning). ApplyConsKeyRotation also
			// no-ops in this case, so skip emitting an update for it.
			if errors.Is(err, types.ErrNoValidatorFound) {
				continue
			}
			return nil, err
		}
		var newPubKey cryptotypes.PubKey
		if err := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); err != nil {
			return nil, err
		}
		pair, err := k.ConsKeyRotationUpdate(validator, newPubKey, powerReduction)
		if err != nil {
			return nil, err
		}
		updates = append(updates, pair...)
	}
	return updates, nil
}

// ConsKeyRotationUpdate builds the (old@0, new@power) ABCI ValidatorUpdate
// pair that announces a cons key rotation to CometBFT. It does not mutate
// state.
func (k Keeper) ConsKeyRotationUpdate(validator types.Validator, newPubKey cryptotypes.PubKey, powerReduction math.Int) (abci.ValidatorUpdates, error) {
	oldTmProtoPk, err := validator.CmtConsPublicKey()
	if err != nil {
		return nil, fmt.Errorf("converting validators existing cons key to tm proto: %w", err)
	}
	newTmProtoPk, err := cryptocodec.ToCmtProtoPublicKey(newPubKey)
	if err != nil {
		return nil, fmt.Errorf("converting validators new cons key to tm proto: %w", err)
	}

	return []abci.ValidatorUpdate{
		{
			PubKey: oldTmProtoPk,
			Power:  0,
		},
		{
			PubKey: newTmProtoPk,
			Power:  validator.ConsensusPower(powerReduction),
		},
	}, nil
}

type PendingRotation struct {
	NewPubKey   cryptotypes.PubKey
	ApplyHeight int64
}

type PendingRotations map[string]PendingRotation

// EffectiveKeyForABCIUpdate returns the consensus pub key that should be used
// for a validator when emitting an ABCI update, given the current set of
// pending rotations.
func (pr PendingRotations) EffectiveKeyForABCIUpdate(valAddr sdk.ValAddress, validator types.Validator) (cryptotypes.PubKey, error) {
	// if this validator has a pending rotation, use the pk that they are
	// rotating to
	if rotation, ok := pr[string(valAddr)]; ok {
		return rotation.NewPubKey, nil
	}

	// the validator is not in the pending rotation set, use their current
	// consensus key
	return validator.ConsPubKey()
}

// EffectiveKeyForGenesis returns the consensus pub key that should be written
// into genesis validator output for an exported initial height. Use this for
// genesis export and InitGenesis validator updates, where a pending rotation
// is only effective if its apply height is at or before the imported chain's
// initial height. Use EffectiveKeyForABCIUpdate instead during end blocker
// validator update emission, where a currently pending rotation should be
// announced with its new key even before the SDK side state swap is applied.
func (pr PendingRotations) EffectiveKeyForGenesis(valAddr sdk.ValAddress, validator types.ValidatorI, exportedInitialHeight int64) (cryptotypes.PubKey, error) {
	// if we have a pending rotation for this validator
	if rotation, ok := pr[string(valAddr)]; ok {
		// if the rotation is going to be applied at a height before the height
		// we restart at, use the pending key as the effective key
		if rotation.ApplyHeight <= exportedInitialHeight {
			return rotation.NewPubKey, nil
		}
	}

	// all other cases use the key currently in state for this val
	return validator.ConsPubKey()
}

// PendingConsKeyRotations scans the apply queue once and returns every
// rotation still in flight, keyed by string(valAddr). It is intended to be
// called once per EndBlock after ProcessConsKeyRotations so the bonded loop
// can substitute the new cons key on per-validator emits via an O(1) map
// lookup instead of repeated store reads.
func (k Keeper) PendingConsKeyRotations(ctx context.Context) (rotations PendingRotations, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(
		types.ConsKeyRotationApplyQueueKey,
		storetypes.PrefixEndBytes(types.ConsKeyRotationApplyQueueKey),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	rotations = make(PendingRotations)
	for ; iterator.Valid(); iterator.Next() {
		applyHeight, valAddr, err := types.ParseConsKeyRotationApplyQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		var newPubKey cryptotypes.PubKey
		if err := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); err != nil {
			return nil, err
		}
		rotations[string(valAddr)] = PendingRotation{
			NewPubKey:   newPubKey,
			ApplyHeight: applyHeight,
		}
	}
	return rotations, nil
}

// PruneMaturedConsKeyRotations removes every rotation whose unbonding window
// has elapsed at the current block time. It deletes the entries from the
// maturity queue, the per validator pending index, and the rotated consensus
// address index.
func (k Keeper) PruneMaturedConsKeyRotations(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockHeader().Time

	keys, err := k.maturedConsKeyRotationKeys(ctx, blockTime)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	for _, key := range keys {
		if err := store.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

// maturedConsKeyRotationKeys walks the maturity queue up to blockTime and
// returns the full set of keys to delete to retire each matured rotation.
func (k Keeper) maturedConsKeyRotationKeys(ctx context.Context, blockTime time.Time) (keys [][]byte, err error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(
		types.ConsKeyRotationQueueKey,
		storetypes.PrefixEndBytes(types.GetConsKeyRotationQueueTimePrefix(blockTime)),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, iterator.Close())
	}()

	// TODO: migrate ValidatorSigningInfo from oldConsAddr to newConsAddr
	for ; iterator.Valid(); iterator.Next() {
		maturity, valAddr, err := types.ParseConsKeyRotationQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		oldConsAddr := sdk.ConsAddress(iterator.Value())

		keys = append(keys,
			types.GetConsKeyRotationQueueKey(maturity, valAddr),
			types.GetValidatorConsKeyRotationKey(valAddr),
			types.GetRotationLockedConsAddrIndexKey(oldConsAddr),
		)
	}
	return keys, nil
}

// rotationApplyHeight returns the height that a rotation should be applied at
// given the current context.
func rotationApplyHeight(ctx context.Context) int64 {
	return sdk.UnwrapSDKContext(ctx).BlockHeight() + types.ConsensusUpdateDelay
}

// rotationMaturityTime returns the time that a rotation will mature at given
// the current context.
func (k Keeper) rotationMaturityTime(ctx context.Context) (time.Time, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	unbondingTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return sdkCtx.BlockHeader().Time.Add(unbondingTime), nil
}
