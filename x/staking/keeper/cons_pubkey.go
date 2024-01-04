package keeper

import (
	"bytes"
	"context"
	"errors"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// maxRotations is the value of max rotations can be made in unbonding period for a validator.
const maxRotations = 1

// setConsPubKeyRotationHistory sets the consensus key rotation of a validator into state
func (k Keeper) setConsPubKeyRotationHistory(
	ctx context.Context, valAddr sdk.ValAddress,
	oldPubKey, newPubKey *codectypes.Any, fee sdk.Coin,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := uint64(sdkCtx.BlockHeight())
	history := types.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr.Bytes(),
		OldConsPubkey:   oldPubKey,
		NewConsPubkey:   newPubKey,
		Height:          height,
		Fee:             fee,
	}
	err := k.RotationHistory.Set(ctx, collections.Join(valAddr.Bytes(), height), history)
	if err != nil {
		return err
	}

	ubdTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return err
	}

	queueTime := sdkCtx.HeaderInfo().Time.Add(ubdTime)
	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Set(ctx, collections.Join(valAddr.Bytes(), queueTime)); err != nil {
		return err
	}

	return k.setConsKeyQueue(ctx, queueTime, valAddr)
}

// updateToNewPubkey gets called from the `ApplyAndReturnValidatorSetUpdates` method during EndBlock.
//
// This method makes the relative state changes to update the keys,
// also maintains a map with old to new conskey rotation which is needed to retrieve the old conskey.
// And also triggers the hook to make changes required in slashing and distribution modules.
func (k Keeper) updateToNewPubkey(ctx context.Context, val types.Validator, oldPubKey, newPubKey *codectypes.Any, fee sdk.Coin) error {
	consAddr, err := val.GetConsAddr()
	if err != nil {
		return err
	}

	if err := k.ValidatorByConsensusAddress.Remove(ctx, consAddr); err != nil {
		return err
	}

	if err := k.DeleteValidatorByPowerIndex(ctx, val); err != nil {
		return err
	}

	val.ConsensusPubkey = newPubKey
	if err := k.SetValidator(ctx, val); err != nil {
		return err
	}
	if err := k.SetValidatorByConsAddr(ctx, val); err != nil {
		return err
	}
	if err := k.SetValidatorByPowerIndex(ctx, val); err != nil {
		return err
	}

	oldPk, ok := oldPubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", oldPk)
	}

	newPk, ok := newPubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", newPk)
	}

	// sets a map: oldConsKey -> newConsKey
	if err := k.OldToNewConsKeyMap.Set(ctx, oldPk.Address(), newPk.Address()); err != nil {
		return err
	}

	// sets a map: newConsKey -> oldConsKey
	if err := k.setNewToOldConsKeyMap(ctx, sdk.ConsAddress(oldPk.Address()), sdk.ConsAddress(newPk.Address())); err != nil {
		return err
	}

	return k.Hooks().AfterConsensusPubKeyUpdate(ctx, oldPk, newPk, fee)
}

// setNewToOldConsKeyMap adds an entry in the state with the current consKey to the initial consKey of the validator.
// it tries to find the oldPk if there is a entry already present in the state
func (k Keeper) setNewToOldConsKeyMap(ctx context.Context, oldPk, newPk sdk.ConsAddress) error {
	pk, err := k.NewToOldConsKeyMap.Get(ctx, oldPk)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	if pk != nil {
		oldPk = pk
	}

	return k.NewToOldConsKeyMap.Set(ctx, newPk, oldPk)
}

// ValidatorIdentifier maps the new cons key to previous cons key (which is the address before the rotation).
// (that is: newConsKey -> oldConsKey)
func (k Keeper) ValidatorIdentifier(ctx context.Context, newPk sdk.ConsAddress) (sdk.ConsAddress, error) {
	pk, err := k.NewToOldConsKeyMap.Get(ctx, newPk)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}

	return pk, nil
}

// exceedsMaxRotations returns true if the key rotations exceed the limit, currently we are limiting one rotation for unbonding period.
func (k Keeper) exceedsMaxRotations(ctx context.Context, valAddr sdk.ValAddress) error {
	count := 0
	rng := collections.NewPrefixedPairRange[[]byte, time.Time](valAddr)

	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Walk(ctx, rng, func(key collections.Pair[[]byte, time.Time]) (stop bool, err error) {
		count++
		return count >= maxRotations, nil
	}); err != nil {
		return err
	}

	if count >= maxRotations {
		return types.ErrExceedingMaxConsPubKeyRotations
	}

	return nil
}

// setConsKeyQueue sets array of rotated validator addresses to a key of current block time + unbonding period
// this is to keep track of rotations made within the unbonding period
func (k Keeper) setConsKeyQueue(ctx context.Context, ts time.Time, valAddr sdk.ValAddress) error {
	queueRec, err := k.ValidatorConsensusKeyRotationRecordQueue.Get(ctx, ts)
	// we should return if the key found here.
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	// push the address if it is not present in the array.
	if !bytesSliceExists(queueRec.Addresses, valAddr.Bytes()) {
		// Address does not exist, so you can append it to the list
		queueRec.Addresses = append(queueRec.Addresses, valAddr.Bytes())
	}

	return k.ValidatorConsensusKeyRotationRecordQueue.Set(ctx, ts, queueRec)
}

// bytesSliceExists tries to find the duplicate entry the array.
func bytesSliceExists(sliceList [][]byte, targetBytes []byte) bool {
	for _, bytesSlice := range sliceList {
		if bytes.Equal(bytesSlice, targetBytes) {
			return true
		}
	}
	return false
}

// PurgeAllMaturedConsKeyRotatedKeys deletes all the matured key rotations.
func (k Keeper) PurgeAllMaturedConsKeyRotatedKeys(ctx sdk.Context, maturedTime time.Time) error {
	maturedRotatedValAddrs, err := k.getAndRemoveAllMaturedRotatedKeys(ctx, maturedTime)
	if err != nil {
		return err
	}

	for _, valAddr := range maturedRotatedValAddrs {
		err := k.deleteConsKeyIndexKey(ctx, valAddr, maturedTime)
		if err != nil {
			return err
		}
	}

	return nil
}

// deleteConsKeyIndexKey deletes the keys which forms a with given validator address and time lesser than the given time.
// eventually there should be only one occurrence since we allow only one rotation for bonding period.
func (k Keeper) deleteConsKeyIndexKey(ctx sdk.Context, valAddr sdk.ValAddress, ts time.Time) error {
	rng := new(collections.Range[collections.Pair[[]byte, time.Time]]).
		StartInclusive(collections.Join(valAddr.Bytes(), time.Time{})).
		EndInclusive(collections.Join(valAddr.Bytes(), ts))

	return k.ValidatorConsensusKeyRotationRecordIndexKey.Walk(ctx, rng, func(key collections.Pair[[]byte, time.Time]) (stop bool, err error) {
		return false, k.ValidatorConsensusKeyRotationRecordIndexKey.Remove(ctx, key)
	})
}

// getAndRemoveAllMaturedRotatedKeys returns all matured valaddresses.
func (k Keeper) getAndRemoveAllMaturedRotatedKeys(ctx sdk.Context, matureTime time.Time) ([][]byte, error) {
	valAddrs := [][]byte{}

	// get an iterator for all timeslices from time 0 until the current HeaderInfo time
	rng := new(collections.Range[time.Time]).EndInclusive(matureTime)
	err := k.ValidatorConsensusKeyRotationRecordQueue.Walk(ctx, rng, func(key time.Time, value types.ValAddrsOfRotatedConsKeys) (stop bool, err error) {
		valAddrs = append(valAddrs, value.Addresses...)
		return false, k.ValidatorConsensusKeyRotationRecordQueue.Remove(ctx, key)
	})
	if err != nil {
		return nil, err
	}

	return valAddrs, nil
}

// GetBlockConsPubKeyRotationHistory returns the rotation history for the current height.
func (k Keeper) GetBlockConsPubKeyRotationHistory(ctx context.Context) ([]types.ConsPubKeyRotationHistory, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	iterator, err := k.RotationHistory.Indexes.Block.MatchExact(ctx, uint64(sdkCtx.BlockHeight()))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	return indexes.CollectValues(ctx, k.RotationHistory, iterator)
}

// GetValidatorConsPubKeyRotationHistory iterates over all the rotated history objects in the state with the given valAddr and returns.
func (k Keeper) GetValidatorConsPubKeyRotationHistory(ctx sdk.Context, operatorAddress sdk.ValAddress) ([]types.ConsPubKeyRotationHistory, error) {
	var historyObjects []types.ConsPubKeyRotationHistory

	rng := collections.NewPrefixedPairRange[[]byte, uint64](operatorAddress.Bytes())

	err := k.RotationHistory.Walk(ctx, rng, func(key collections.Pair[[]byte, uint64], history types.ConsPubKeyRotationHistory) (stop bool, err error) {
		historyObjects = append(historyObjects, history)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return historyObjects, nil
}
