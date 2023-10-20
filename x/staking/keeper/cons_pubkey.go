package keeper

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/collections"
)

// SetConsPubKeyRotationHistory sets the consensus key rotation of a validator into state
func (k Keeper) SetConsPubKeyRotationHistory(
	ctx sdk.Context, valAddr sdk.ValAddress,
	oldPubKey, newPubKey *codectypes.Any, height uint64, fee sdk.Coin,
) error {
	history := types.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr.String(),
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

	queueTime := ctx.BlockHeader().Time.Add(ubdTime)
	k.ValidatorConsensusKeyRotationRecordIndexKey.Set(ctx, collections.Join(valAddr.Bytes(), queueTime), []byte{})
	k.SetConsKeyQueue(ctx, queueTime, valAddr)

	return nil
}

// CheckLimitOfMaxRotationsExceed returns bool, count of iterations made within the unbonding period.
func (k Keeper) CheckLimitOfMaxRotationsExceed(ctx sdk.Context, valAddr sdk.ValAddress) bool {
	isFound := false
	rng := collections.NewPrefixUntilPairRange[[]byte, time.Time](valAddr)
	k.ValidatorConsensusKeyRotationRecordIndexKey.Walk(ctx, rng, func(key collections.Pair[[]byte, time.Time], value []byte) (stop bool, err error) {
		isFound = true
		return true, nil
	})

	return isFound
}

// SetConsKeyQueue sets array of rotated validator addresses to a key of current block time + unbonding period
// this is to keep track of rotations made within the unbonding period
func (k Keeper) SetConsKeyQueue(ctx sdk.Context, ts time.Time, valAddr sdk.ValAddress) error {
	queueRec, err := k.ValidatorConsensusKeyRotationRecordQueue.Get(ctx, ts)
	if err != nil {
		return err
	}

	queueRec.Addresses = append(queueRec.Addresses, valAddr.String())
	if err := k.ValidatorConsensusKeyRotationRecordQueue.Set(ctx, ts, queueRec); err != nil {
		return err
	}

	return nil
}
