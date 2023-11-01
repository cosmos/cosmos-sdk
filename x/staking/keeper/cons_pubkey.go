package keeper

import (
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Set(ctx, collections.Join(valAddr.Bytes(), queueTime), []byte{}); err != nil {
		return err
	}

	return k.SetConsKeyQueue(ctx, queueTime, valAddr)
}

func (k Keeper) updateToNewPubkey(ctx sdk.Context, val types.Validator, oldPubKey, newPubKey *codectypes.Any, fee sdk.Coin) error {
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

// CheckLimitOfMaxRotationsExceed returns bool, count of iterations made within the unbonding period.
// CheckLimitOfMaxRotationsExceed returns true if the key rotations exceed the limit, currently we are limiting one rotation for unbonding period.
func (k Keeper) CheckLimitOfMaxRotationsExceed(ctx sdk.Context, valAddr sdk.ValAddress) (bool, error) {
	count := 0
	maxRotations := 1 // arbitrary value
	rng := collections.NewPrefixUntilPairRange[[]byte, time.Time](valAddr)
	if err := k.ValidatorConsensusKeyRotationRecordIndexKey.Walk(ctx, rng, func(key collections.Pair[[]byte, time.Time], value []byte) (stop bool, err error) {
		count++
		if count >= maxRotations {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return false, err
	}

	return count >= maxRotations, nil
}

// SetConsKeyQueue sets array of rotated validator addresses to a key of current block time + unbonding period
// this is to keep track of rotations made within the unbonding period
func (k Keeper) SetConsKeyQueue(ctx sdk.Context, ts time.Time, valAddr sdk.ValAddress) error {
	queueRec, err := k.ValidatorConsensusKeyRotationRecordQueue.Get(ctx, ts)
	if err != nil {
		return err
	}

	queueRec.Addresses = append(queueRec.Addresses, valAddr.String())
	return k.ValidatorConsensusKeyRotationRecordQueue.Set(ctx, ts, queueRec)
}

// GetValidatorConsPubKeyRotationHistory iterates over all the rotated history objects in the state with the given valAddr and returns.
func (k Keeper) GetValidatorConsPubKeyRotationHistory(ctx sdk.Context, valAddr sdk.ValAddress) ([]types.ConsPubKeyRotationHistory, error) {
	var historyObjects []types.ConsPubKeyRotationHistory
	rng := collections.NewPrefixUntilPairRange[[]byte, uint64](valAddr)
	err := k.ValidatorConsPubKeyRotationHistory.Walk(ctx, rng, func(key collections.Pair[[]byte, uint64], value types.ConsPubKeyRotationHistory) (stop bool, err error) {
		historyObjects = append(historyObjects, value)
		return
	})
	if err != nil {
		return nil, err
	}

	return historyObjects, nil
}

// GetBlockConsPubKeyRotationHistory iterator over the rotation history for the given height.
func (k Keeper) GetBlockConsPubKeyRotationHistory(ctx sdk.Context, height uint64) ([]types.ConsPubKeyRotationHistory, error) {
	var historyObjects []types.ConsPubKeyRotationHistory

	rng := new(collections.Range[uint64]).Prefix(height)
	err := k.BlockConsPubKeyRotationHistory.Walk(ctx, rng, func(key uint64, value types.ConsPubKeyRotationHistory) (stop bool, err error) {
		historyObjects = append(historyObjects, value)
		return
	})

	if err != nil {
		return nil, err
	}

	return historyObjects, nil
}
