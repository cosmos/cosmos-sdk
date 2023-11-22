package keeper

import (
	"bytes"
	"context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	err := k.RotationHistory.Set(ctx, valAddr, history)
	if err != nil {
		return err
	}

	ubdTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return err
	}

	queueTime := sdkCtx.BlockHeader().Time.Add(ubdTime)
	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Set(ctx, collections.Join(valAddr.Bytes(), queueTime)); err != nil {
		return err
	}

	return k.setConsKeyQueue(ctx, queueTime, valAddr)
}

// This method gets called from the `ApplyAndReturnValidatorSetUpdates`(from endblocker) method.
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

	oldPk := oldPubKey.GetCachedValue().(cryptotypes.PubKey)
	newPk := newPubKey.GetCachedValue().(cryptotypes.PubKey)

	// Sets a map to newly rotated consensus key with old consensus key
	if err := k.RotatedConsKeyMapIndex.Set(ctx, oldPk.Address(), newPk.Address()); err != nil {
		return err
	}

	return k.Hooks().AfterConsensusPubKeyUpdate(ctx, oldPk, newPk, fee)
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
	if err != nil {
		return err
	}

	if !bytesSliceExists(queueRec.Addresses, valAddr.Bytes()) {
		// Address does not exist, so you can append it to the list
		queueRec.Addresses = append(queueRec.Addresses, valAddr.Bytes())
	}

	return k.ValidatorConsensusKeyRotationRecordQueue.Set(ctx, ts, queueRec)
}

func bytesSliceExists(sliceList [][]byte, targetBytes []byte) bool {
	for _, bytesSlice := range sliceList {
		if bytes.Equal(bytesSlice, targetBytes) {
			return true
		}
	}
	return false
}

// UpdateAllMaturedConsKeyRotatedKeys udpates all the matured key rotations.
func (k Keeper) UpdateAllMaturedConsKeyRotatedKeys(ctx sdk.Context, maturedTime time.Time) error {
	maturedRotatedValAddrs, err := k.GetAllMaturedRotatedKeys(ctx, maturedTime)
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

// deleteConsKeyIndexKey deletes the key which is formed with the given valAddr, time.
func (k Keeper) deleteConsKeyIndexKey(ctx sdk.Context, valAddr sdk.ValAddress, ts time.Time) error {
	rng := new(collections.Range[collections.Pair[[]byte, time.Time]]).
		EndInclusive(collections.Join(valAddr.Bytes(), ts))

	return k.ValidatorConsensusKeyRotationRecordIndexKey.Walk(ctx, rng, func(key collections.Pair[[]byte, time.Time]) (stop bool, err error) {
		return false, k.ValidatorConsensusKeyRotationRecordIndexKey.Remove(ctx, key)
	})
}

// GetAllMaturedRotatedKeys returns all matured valaddresses .
func (k Keeper) GetAllMaturedRotatedKeys(ctx sdk.Context, matureTime time.Time) ([][]byte, error) {
	valAddrs := [][]byte{}

	// get an iterator for all timeslices from time 0 until the current HeaderInfo time
	rng := new(collections.Range[time.Time]).EndInclusive(matureTime)
	iterator, err := k.ValidatorConsensusKeyRotationRecordQueue.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	// indexes.CollectValues(ctx, k.vaValidatorConsensusKeyRotationRecordQueue, iterator)

	for ; iterator.Valid(); iterator.Next() {
		value, err := iterator.Value()
		if err != nil {
			return nil, err
		}
		valAddrs = append(valAddrs, value.Addresses...)
		key, err := iterator.Key()
		if err != nil {
			return nil, err
		}

		err = k.ValidatorConsensusKeyRotationRecordQueue.Remove(ctx, key)
		if err != nil {
			return nil, err
		}
	}

	return valAddrs, nil
}

// GetBlockConsPubKeyRotationHistory iterator over the rotation history for the given height.
func (k Keeper) GetBlockConsPubKeyRotationHistory(ctx context.Context) ([]types.ConsPubKeyRotationHistory, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	var historyObjects []types.ConsPubKeyRotationHistory

	iterator, err := k.RotationHistory.Indexes.Block.MatchExact(ctx, uint64(sdkCtx.BlockHeight()))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	keys, err := iterator.PrimaryKeys()
	if err != nil {
		return nil, err
	}

	for _, v := range keys {
		history, err := k.RotationHistory.Get(ctx, v)
		if err != nil {
			return nil, err
		}

		historyObjects = append(historyObjects, history)
	}

	return historyObjects, nil
}
