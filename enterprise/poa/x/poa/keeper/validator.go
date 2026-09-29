// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/collections"
	sdkerrors "cosmossdk.io/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UpdateValidator updates a single validator's power and metadata.
// It validates the power is non-negative, updates the validator state, and queues an ABCI update if power changed.
func (k *Keeper) UpdateValidator(ctx sdk.Context, consAddress sdk.ConsAddress, updates types.Validator) error {
	if updates.Power < 0 {
		return types.ErrNegativeValidatorPower
	}

	existingValidator, err := k.validators.Get(ctx, consAddress)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrUnknownValidator
		}
		return err
	}

	oldPower := existingValidator.Power

	// Apply updates
	existingValidator.Power = updates.Power
	if updates.Metadata != nil {
		pubKeyToCheck := existingValidator.PubKey
		if updates.PubKey != nil {
			pubKeyToCheck = updates.PubKey
		}

		if err := k.ValidateOperatorAndConsensusPubKeyDifferent(updates.Metadata.OperatorAddress, pubKeyToCheck); err != nil {
			return err
		}

		existingValidator.Metadata = updates.Metadata
	}
	if updates.PubKey != nil {
		existingValidator.PubKey = updates.PubKey
	}

	if oldPower != updates.Power {
		if err := k.checkpointAllValidators(ctx); err != nil {
			return err
		}
		if err := k.AdjustTotalPower(ctx, updates.Power-oldPower); err != nil {
			return err
		}
	}

	if err := k.validators.Set(ctx, consAddress, existingValidator); err != nil {
		return err
	}

	// Create validator update for CometBFT only if power actually changed
	if oldPower != updates.Power {
		update, err := k.createABCIValidatorUpdate(existingValidator.PubKey, existingValidator.Power)
		if err != nil {
			return err
		}
		if err := k.queuedUpdates.Push(ctx, update); err != nil {
			return err
		}
	}

	return nil
}

// UpdateValidators updates multiple validators in a single operation.
// The operation is atomic: if any validator update fails, all changes are reverted.
func (k *Keeper) UpdateValidators(ctx sdk.Context, validators []types.Validator) error {
	cacheCtx, writeCache := ctx.CacheContext()

	for _, validator := range validators {
		var pubKey cryptotypes.PubKey
		if err := k.cdc.UnpackAny(validator.PubKey, &pubKey); err != nil {
			return err
		}
		if err := k.validatePubkeyType(cacheCtx, pubKey); err != nil {
			return err
		}

		consAddress := sdk.GetConsAddress(pubKey)

		if err := k.UpdateValidator(cacheCtx, consAddress, validator); err != nil {
			return err
		}
	}

	writeCache()
	return nil
}

// CreateValidator creates a new validator with the specified initial state.
// The validator can be created with non-zero power if needed (e.g., during genesis).
func (k *Keeper) CreateValidator(ctx sdk.Context, consAddress sdk.ConsAddress, validator types.Validator, checkpoint bool) error {
	exists, err := k.validators.Has(ctx, consAddress)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrValidatorAlreadyExists
	}

	if err := k.ValidateOperatorAndConsensusPubKeyDifferent(validator.Metadata.OperatorAddress, validator.PubKey); err != nil {
		return err
	}

	if validator.Power > 0 {
		if checkpoint {
			if err := k.checkpointAllValidators(ctx); err != nil {
				return err
			}
		}
		update, err := k.createABCIValidatorUpdate(validator.PubKey, validator.Power)
		if err != nil {
			return err
		}
		if err := k.queuedUpdates.Push(ctx, update); err != nil {
			return err
		}

		if err := k.AdjustTotalPower(ctx, validator.Power); err != nil {
			return err
		}
	}

	return k.validators.Set(ctx, consAddress, validator)
}

// RotateConsPubKey rotates a validator's consensus pubkey in place, re-keying
// the record from its current consensus address to the one derived from
// newPubKey, preserving power and metadata.
func (k *Keeper) RotateConsPubKey(ctx sdk.Context, operatorAddr sdk.AccAddress, newPubKey cryptotypes.PubKey) error {
	// ensure validator being rotated exists
	validator, err := k.GetValidatorByOperatorAddress(ctx, operatorAddr)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrUnknownValidator
		}
		return err
	}

	// ensure pubkey type is allowed
	if err := k.validatePubkeyType(ctx, newPubKey); err != nil {
		return err
	}

	var oldPubKey cryptotypes.PubKey
	if err := k.cdc.UnpackAny(validator.PubKey, &oldPubKey); err != nil {
		return err
	}
	oldConsAddr := sdk.GetConsAddress(oldPubKey)
	newConsAddr := sdk.GetConsAddress(newPubKey)

	// disallow rotations to the same key
	if oldConsAddr.Equals(newConsAddr) {
		return types.ErrNoOpRotation
	}

	// disallow rotations to a key another validator is currently using
	inUse, err := k.validators.Has(ctx, newConsAddr)
	if err != nil {
		return err
	}
	if inUse {
		return types.ErrConsensusPubKeyInUse
	}

	newPubKeyAny, err := codectypes.NewAnyWithValue(newPubKey)
	if err != nil {
		return err
	}

	// disallow setting the consensus key to your key backing the operator
	// address
	if err := k.ValidateOperatorAndConsensusPubKeyDifferent(validator.Metadata.OperatorAddress, newPubKeyAny); err != nil {
		return err
	}

	// settle pending fee accruals under the old consensus address before
	// rotating. otherwise the rotation moves the record before the outstanding
	// share lands in the old cons addr's bucket and it would be stranded
	if err := k.checkpointAllValidators(ctx); err != nil {
		return err
	}

	oldPubKeyAny := validator.PubKey
	power := validator.Power
	validator.PubKey = newPubKeyAny

	// do the rotation
	if err := k.validators.Remove(ctx, oldConsAddr); err != nil {
		return err
	}
	if err := k.validators.Set(ctx, newConsAddr, validator); err != nil {
		return err
	}

	// migrate accrued fees from the old cons addr to the new one so the
	// operator's balance follows the rotation without a forced payout
	if err := k.migrateAllocatedFees(ctx, oldConsAddr, newConsAddr); err != nil {
		return err
	}

	// queue abci updates only when the validator is in CometBFT's active set.
	//
	// NOTE: a 0 power validator is not in the set, and comet treats a 0 power
	// update as a deletion of an address not in the set, which aborts applying
	// the block.
	if power > 0 {
		// create and push an update to set the old key to 0 power, removing it
		removeOld, err := k.createABCIValidatorUpdate(oldPubKeyAny, 0)
		if err != nil {
			return err
		}
		if err := k.queuedUpdates.Push(ctx, removeOld); err != nil {
			return err
		}

		// create and push an update to set the new key to the old keys power,
		// activating it
		addNew, err := k.createABCIValidatorUpdate(newPubKeyAny, power)
		if err != nil {
			return err
		}
		if err := k.queuedUpdates.Push(ctx, addNew); err != nil {
			return err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRotateConsPubKey,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, operatorAddr.String()),
			sdk.NewAttribute(types.AttributeKeyConsensusAddress, oldConsAddr.String()),
			sdk.NewAttribute(types.AttributeKeyNewConsensusAddress, newConsAddr.String()),
		),
	)

	return nil
}

// GetValidatorByOperatorAddress retrieves a validator by operator address using the secondary index.
func (k *Keeper) GetValidatorByOperatorAddress(ctx sdk.Context, operatorAddr sdk.AccAddress) (types.Validator, error) {
	consAddr, err := k.validators.Indexes.OperatorAddress.MatchExact(ctx, operatorAddr.String())
	if err != nil {
		return types.Validator{}, err
	}
	return k.validators.Get(ctx, consAddr)
}

// createABCIValidatorUpdate creates a CometBFT validator update from a
// validator's public key and power.
func (k *Keeper) createABCIValidatorUpdate(pubKeyAny *codectypes.Any, power int64) (abci.ValidatorUpdate, error) {
	var pubKey cryptotypes.PubKey
	if err := k.cdc.UnpackAny(pubKeyAny, &pubKey); err != nil {
		return abci.ValidatorUpdate{}, err
	}

	cmtPk, err := codec.ToCmtProtoPublicKey(pubKey)
	if err != nil {
		return abci.ValidatorUpdate{}, err
	}

	return abci.ValidatorUpdate{PubKey: cmtPk, Power: power}, nil
}

// IterateActiveValidators walks the power index in descending order, skipping validators with power 0.
func (k *Keeper) IterateActiveValidators(ctx sdk.Context, callback func(consAddr sdk.ConsAddress, power int64, validator types.Validator) (stop bool, err error)) error {
	ranger := new(collections.Range[collections.Pair[int64, sdk.ConsAddress]]).Descending()
	return k.validators.Indexes.Power.Walk(ctx, ranger, func(power int64, consAddr sdk.ConsAddress) (bool, error) {
		if power == 0 {
			return true, nil
		}
		validator, err := k.validators.Get(ctx, consAddr)
		if err != nil {
			return true, err
		}
		return callback(consAddr, power, validator)
	})
}

// GetTotalPower returns the total validator power across all active validators.
func (k *Keeper) GetTotalPower(ctx sdk.Context) (int64, error) {
	power, err := k.totalPower.Get(ctx)
	if err != nil {
		if ctx.BlockHeight() == 0 {
			return 0, nil
		} else {
			return 0, err
		}
	}
	return power, nil
}

// AdjustTotalPower adjusts the total power by the given delta.
// It ensures the total power never becomes zero or negative.
func (k *Keeper) AdjustTotalPower(ctx sdk.Context, delta int64) error {
	if delta == 0 {
		return nil
	}

	current, err := k.GetTotalPower(ctx)
	if err != nil {
		return err
	}

	newTotal := current + delta
	if newTotal < 0 {
		return sdkerrors.Wrapf(types.ErrInvalidTotalPower, "total power would become negative: %d", newTotal)
	}

	if newTotal == 0 {
		return sdkerrors.Wrap(types.ErrInvalidTotalPower, "total power cannot be zero")
	}

	return k.totalPower.Set(ctx, newTotal)
}

// GetAllValidators returns all validators in descending power order.
func (k *Keeper) GetAllValidators(ctx sdk.Context) ([]types.Validator, error) {
	var validators []types.Validator
	ranger := new(collections.Range[collections.Pair[int64, sdk.ConsAddress]]).Descending()
	err := k.validators.Indexes.Power.Walk(ctx, ranger, func(power int64, consAddr sdk.ConsAddress) (bool, error) {
		validator, err := k.validators.Get(ctx, consAddr)
		if err != nil {
			return true, err
		}
		validators = append(validators, validator)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return validators, nil
}
