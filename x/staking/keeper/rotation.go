package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"

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
			// if we import the genesis without modifying the apply height,
			// comet will not know about this key rotation and we will update
			// sdk state without updating comet. thus, we push the apply
			// height forward so that the abci updates can be reemitted and
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

// SetRotationLockedConsAddr sets the lock record for a consensus address used
// in a key rotation.
func (k Keeper) SetRotationLockedConsAddr(
	ctx context.Context,
	consAddr sdk.ConsAddress,
	valAddr sdk.ValAddress,
	kind types.ConsAddrLockType,
) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Set(types.GetRotationLockedConsAddrIndexKey(consAddr), types.RotationLockedConsAddrIndexValue(kind, valAddr))
}

// GetRotationLockedConsAddr returns the lock record for a consensus address
// used in a key rotation.
func (k Keeper) GetRotationLockedConsAddr(
	ctx context.Context,
	consAddr sdk.ConsAddress,
) (kind types.ConsAddrLockType, valAddr sdk.ValAddress, found bool, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetRotationLockedConsAddrIndexKey(consAddr))
	if err != nil {
		return 0, nil, false, err
	}
	if bz == nil {
		return 0, nil, false, nil
	}

	kind, valAddr, err = types.ParseRotationLockedConsAddrIndexValue(bz)
	if err != nil {
		return 0, nil, false, err
	}
	return kind, valAddr, true, nil
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

	// Lock both the old and new cons addrs so that no validator can rotate to
	// either while the rotation is pending. PendingFrom protects the old addr
	// if the validator is removed before apply, but is not historical evidence
	// lookup state. PendingTo reserves the future addr. ApplyConsKeyRotation
	// changes the old lock to RotatedFrom after the validator's live cons addr
	// index moves, and ValidatorByHistoricalConsAddr only resolves that kind.
	if err := k.SetRotationLockedConsAddr(ctx, oldConsAddr, valAddr, types.ConsAddrLockPendingFrom); err != nil {
		return err
	}
	if err := k.SetRotationLockedConsAddr(ctx, newConsAddr, valAddr, types.ConsAddrLockPendingTo); err != nil {
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
		if err := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); err != nil {
			return err
		}
		entries = append(entries, matured{key: keyCopy, valAddr: valAddr, newPubKey: newPubKey})
	}

	for _, e := range entries {
		if err := k.ApplyConsKeyRotation(ctx, e.valAddr, e.newPubKey); err != nil {
			return err
		}

		// the new cons addr is now the validator's live cons addr. further
		// rotations targeting it are blocked by the by cons addr lookup, so
		// release its rotation lock entry. The old cons addr entry stays
		// until the rotation matures.
		if err := store.Delete(types.GetRotationLockedConsAddrIndexKey(sdk.ConsAddress(e.newPubKey.Address()))); err != nil {
			return err
		}

		// delete the entry from the apply queue
		if err := store.Delete(e.key); err != nil {
			return err
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
	if err := k.SetValidatorByConsAddr(ctx, validator); err != nil {
		return err
	}

	newConsAddr := sdk.ConsAddress(newPubKey.Address())
	if err := k.SetRotationLockedConsAddr(ctx, oldConsAddr, valAddr, types.ConsAddrLockRotatedFrom); err != nil {
		return err
	}
	return k.Hooks().AfterValidatorConsKeyUpdated(ctx, oldConsAddr, newConsAddr, valAddr)
}

// PendingConsKeyRotationUpdate stores the rotation metadata needed to align
// staking-side validator updates with CometBFT's delayed key-rotation view.
type PendingConsKeyRotationUpdate struct {
	// OldPubKey is the consensus key currently stored in SDK validator state.
	OldPubKey cryptotypes.PubKey

	// NewPubKey is the pending consensus key CometBFT should track for the
	// validator once the rotation update is emitted.
	NewPubKey cryptotypes.PubKey

	// EmitHeight is the EndBlock height that should emit the old@0,new@power
	// rotation pair.
	EmitHeight int64

	// LastPower is the validator power from the last Comet-visible validator
	// set, keyed by operator address.
	LastPower int64
}

// oldAddr returns the Comet validator address for the pre-rotation key.
func (r PendingConsKeyRotationUpdate) oldAddr() string {
	return string(r.OldPubKey.Address())
}

// shouldEmitPowerUpdates returns whether the rotation should emit
// old@0,new@power at height.
func (r PendingConsKeyRotationUpdate) shouldEmitPowerUpdates(height int64) bool {
	return r.EmitHeight == height && r.LastPower > 0
}

// PendingConsKeyRotationUpdates returns rotation metadata for in-flight
// rotations whose CometBFT update has been or should be emitted by the current
// EndBlock. It must be called after ApplyConsKeyRotations so matured entries
// have already been drained.
func (k Keeper) PendingConsKeyRotationUpdates(ctx context.Context, last map[string]int64) (updates []PendingConsKeyRotationUpdate, err error) {
	store := k.storeService.OpenKVStore(ctx)
	currentHeight := sdk.UnwrapSDKContext(ctx).BlockHeight()

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

	for ; iterator.Valid(); iterator.Next() {
		applyHeight, valAddr, err := types.ParseConsKeyRotationApplyQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}
		emitHeight := applyHeight - types.ConsensusUpdateDelay
		// Future rotations are not visible to Comet yet. Older entries still
		// matter until applyHeight because normal staking updates must be
		// translated from the stored old key to the Comet-visible new key.
		if emitHeight > currentHeight {
			continue
		}

		validator, err := k.GetValidator(ctx, valAddr)
		if err != nil {
			if errors.Is(err, types.ErrNoValidatorFound) {
				continue
			}
			return nil, err
		}
		oldPubKey, err := validator.ConsPubKey()
		if err != nil {
			return nil, err
		}
		var newPubKey cryptotypes.PubKey
		if err := k.cdc.UnmarshalInterface(iterator.Value(), &newPubKey); err != nil {
			return nil, err
		}

		lastPower := last[string(valAddr)]
		updates = append(updates, PendingConsKeyRotationUpdate{
			OldPubKey:  oldPubKey,
			NewPubKey:  newPubKey,
			EmitHeight: emitHeight,
			LastPower:  lastPower,
		})
	}
	return updates, nil
}

// ProcessValidatorUpdatesForConsKeyRotations rewrites validator updates that
// reference old consensus keys so the returned batch matches CometBFT's
// key-rotation timeline.
func (k Keeper) ProcessValidatorUpdatesForConsKeyRotations(
	ctx context.Context,
	rotations []PendingConsKeyRotationUpdate,
	updates []abci.ValidatorUpdate,
) ([]abci.ValidatorUpdate, error) {
	if len(rotations) == 0 {
		return updates, nil
	}

	currentHeight := sdk.UnwrapSDKContext(ctx).BlockHeight()
	updateSet, err := newValidatorUpdateSet(updates)
	if err != nil {
		return nil, err
	}

	for _, rotation := range rotations {
		// check if this rotating validator has emitted power updates on its
		// old consensus address. if it has, then we may need to rewrite those
		// to be associated with the cons addr we are rotating to.
		oldAddr := rotation.oldAddr()
		if update, ok := updateSet.Get(oldAddr); ok {
			// depending on what the validator power update on the old cons
			// addr looks like, modify taking into account the pending key
			// rotation.
			rewritten, err := rotatedValidatorUpdates(update, rotation, currentHeight)
			if err != nil {
				return nil, err
			}
			updateSet.Replace(oldAddr, rewritten...)
			continue
		}

		// the rotating validator does not have any power updates being emitted
		// at this height

		// determine if the rotating validator should emit power updates based
		// on its current power and the designated emit height to signal to
		// comet that the validator is changing
		if !rotation.shouldEmitPowerUpdates(currentHeight) {
			continue
		}

		// we should emit power updates for this validator at this height

		// set the key they are rotating away from to 0
		oldUpdate, err := validatorUpdateForPubKey(rotation.OldPubKey, 0)
		if err != nil {
			return nil, err
		}

		// and set the new key they are rotating to to the validators last seen
		// power
		newUpdate, err := validatorUpdateForPubKey(rotation.NewPubKey, rotation.LastPower)
		if err != nil {
			return nil, err
		}
		updateSet.Append([]abci.ValidatorUpdate{oldUpdate, newUpdate}...)
	}

	return updateSet.Updates()
}

// rotatedValidatorUpdates rewrites a normal staking update for a validator
// whose consensus key rotation is still in flight.
func rotatedValidatorUpdates(
	update abci.ValidatorUpdate,
	rotation PendingConsKeyRotationUpdate,
	currentHeight int64,
) ([]abci.ValidatorUpdate, error) {
	switch {
	case update.Power == 0 && rotation.shouldEmitPowerUpdates(currentHeight):
		// A same-block status/power transition removed the validator before the
		// new key should be added. Keep oldkey@0 and avoid emitting newkey@0 for a key
		// Comet has never seen.
		return []abci.ValidatorUpdate{update}, nil
	case update.Power == 0:
		// The rotation was already emitted in an earlier EndBlock, so Comet now
		// tracks the new key even though SDK state still stores the old key.
		newUpdate, err := validatorUpdateForPubKey(rotation.NewPubKey, 0)
		if err != nil {
			return nil, err
		}
		return []abci.ValidatorUpdate{newUpdate}, nil
	case rotation.shouldEmitPowerUpdates(currentHeight):
		// Normal staking emitted old@power at the same height the rotation pair
		// is due. Convert it into the Comet key swap old@0, new@power.
		newUpdate, err := validatorUpdateForPubKey(rotation.NewPubKey, update.Power)
		if err != nil {
			return nil, err
		}
		return []abci.ValidatorUpdate{
			{PubKey: update.PubKey, Power: 0},
			newUpdate,
		}, nil
	default:
		// The key swap was already emitted to Comet, but the SDK-side key swap
		// has not reached applyHeight. Translate old@power to new@power.
		newUpdate, err := validatorUpdateForPubKey(rotation.NewPubKey, update.Power)
		if err != nil {
			return nil, err
		}
		return []abci.ValidatorUpdate{newUpdate}, nil
	}
}

// validatorUpdateForPubKey builds an ABCI validator update for a pubkey.
func validatorUpdateForPubKey(pk cryptotypes.PubKey, power int64) (abci.ValidatorUpdate, error) {
	cmtPubKey, err := cryptocodec.ToCmtProtoPublicKey(pk)
	if err != nil {
		return abci.ValidatorUpdate{}, err
	}
	return abci.ValidatorUpdate{PubKey: cmtPubKey, Power: power}, nil
}

// validatorUpdateSet stores normal validator updates by consensus address while
// preserving their original order and allowing rotation rewrites to replace
// one normal update with zero, one, or multiple updates.
type validatorUpdateSet struct {
	// order preserves the Comet consensus address order of the original
	// validator-state updates.
	order []string

	// replacements stores the update or replacement updates for each original
	// consensus address.
	replacements map[string][]abci.ValidatorUpdate

	// extra stores updates introduced by rotations that had no original
	// validator-state update in this EndBlock.
	extra []abci.ValidatorUpdate
}

// newValidatorUpdateSet creates a new validator update set in order to index
// validator updates by cons addr.
func newValidatorUpdateSet(updates []abci.ValidatorUpdate) (*validatorUpdateSet, error) {
	updateSet := &validatorUpdateSet{
		order:        make([]string, 0, len(updates)),
		replacements: make(map[string][]abci.ValidatorUpdate, len(updates)),
	}
	for _, update := range updates {
		addr, err := validatorUpdateAddress(update)
		if err != nil {
			return nil, err
		}
		if _, ok := updateSet.replacements[addr]; ok {
			return nil, fmt.Errorf("duplicate validator update for consensus address %X", []byte(addr))
		}
		updateSet.order = append(updateSet.order, addr)
		updateSet.replacements[addr] = []abci.ValidatorUpdate{update}
	}
	return updateSet, nil
}

// Get returns the first update currently associated with addr.
func (s *validatorUpdateSet) Get(addr string) (abci.ValidatorUpdate, bool) {
	updates, ok := s.replacements[addr]
	if !ok || len(updates) == 0 {
		return abci.ValidatorUpdate{}, false
	}
	return updates[0], true
}

// Replace replaces the updates associated with addr.
func (s *validatorUpdateSet) Replace(addr string, updates ...abci.ValidatorUpdate) {
	s.replacements[addr] = updates
}

// Append adds updates after all originally ordered updates.
func (s *validatorUpdateSet) Append(updates ...abci.ValidatorUpdate) {
	s.extra = append(s.extra, updates...)
}

// Updates returns the ordered validator updates and rejects duplicate addresses.
func (s *validatorUpdateSet) Updates() ([]abci.ValidatorUpdate, error) {
	updates := make([]abci.ValidatorUpdate, 0, len(s.order)+len(s.extra))
	for _, addr := range s.order {
		updates = append(updates, s.replacements[addr]...)
	}
	updates = append(updates, s.extra...)
	return updates, nil
}

// validatorUpdateAddress returns the cons addr for a validator update.
func validatorUpdateAddress(update abci.ValidatorUpdate) (string, error) {
	pk, err := cryptoenc.PubKeyFromProto(update.PubKey)
	if err != nil {
		return "", err
	}
	return string(pk.Address()), nil
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
