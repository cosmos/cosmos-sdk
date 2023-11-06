package keeper

import (
	"bytes"
	"context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	err := k.ValidatorConsPubKeyRotationHistory.Set(ctx, collections.Join(valAddr.Bytes(), height), history)
	if err != nil {
		return err
	}

	if err := k.BlockConsPubKeyRotationHistory.Set(ctx, height, history); err != nil {
		return err
	}

	ubdTime, err := k.UnbondingTime(ctx)
	if err != nil {
		return err
	}

	queueTime := sdkCtx.BlockHeader().Time.Add(ubdTime)
	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Set(ctx, collections.Join(valAddr.Bytes(), queueTime), []byte{}); err != nil {
		return err
	}

	return k.SetConsKeyQueue(ctx, queueTime, valAddr)
}

// exceedsMaxRotations returns true if the key rotations exceed the limit, currently we are limiting one rotation for unbonding period.
func (k Keeper) exceedsMaxRotations(ctx context.Context, valAddr sdk.ValAddress) (bool, error) {
	count := 0
	rng := collections.NewPrefixedPairRange[[]byte, time.Time](valAddr)
	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Walk(ctx, rng, func(key collections.Pair[[]byte, time.Time], value []byte) (stop bool, err error) {
		count++
		return count >= maxRotations, nil
	}); err != nil {
		return false, err
	}

	return count >= maxRotations, nil
}

// SetConsKeyQueue sets array of rotated validator addresses to a key of current block time + unbonding period
// this is to keep track of rotations made within the unbonding period
func (k Keeper) SetConsKeyQueue(ctx context.Context, ts time.Time, valAddr sdk.ValAddress) error {
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
