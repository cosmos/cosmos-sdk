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
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/collections"
	sdkerrors "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

// UpdateValidator updates a single validator's power and metadata.
// It validates the power is non-negative, updates the validator state, and queues an ABCI update if power changed.
func (k *Keeper) UpdateValidator(ctx sdk.Context, consAddress sdk.ConsAddress, updates types.Validator) error {
	// Validate power
	if updates.Power < 0 {
		return types.ErrNegativeValidatorPower
	}

	// Check validator exists
	if ok, err := k.HasValidator(ctx, consAddress); err != nil {
		return err
	} else if !ok {
		return types.ErrUnknownValidator
	}

	// Get current state
	oldPower, err := k.GetValidatorPower(ctx, consAddress)
	if err != nil {
		return err
	}

	existingValidator, err := k.GetValidator(ctx, consAddress)
	if err != nil {
		return err
	}

	// Apply updates
	existingValidator.Power = updates.Power
	if updates.Metadata != nil {
		// Get the pubkey to check (use updated one if provided, otherwise existing)
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

	// Delete old entry and insert with new power
	// Handles the total validator power.
	if err := k.SetValidatorPower(ctx, consAddress, existingValidator.Power); err != nil {
		return err
	}

	// Update the full validator object (SetValidatorPower only updates power)
	compositeKey := collections.Join(existingValidator.Power, consAddress.String())
	if err := k.validators.Set(ctx, compositeKey, existingValidator); err != nil {
		return err
	}

	// Create validator update for CometBFT
	// Only send update if power actually changed
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
func (k *Keeper) UpdateValidators(ctx sdk.Context, validators []types.Validator) error {
	for _, validator := range validators {
		var pubKey cryptotypes.PubKey
		if err := k.cdc.UnpackAny(validator.PubKey, &pubKey); err != nil {
			return err
		}
		if err := k.validatePubkeyType(ctx, pubKey); err != nil {
			return err
		}

		consAddress := sdk.GetConsAddress(pubKey)

		if err := k.UpdateValidator(ctx, consAddress, validator); err != nil {
			return err
		}
	}

	return nil
}

// CreateValidator creates a new validator with the specified initial state.
// The validator can be created with non-zero power if needed (e.g., during genesis).
func (k *Keeper) CreateValidator(ctx sdk.Context, consAddress sdk.ConsAddress, validator types.Validator, checkpoint bool) error {
	// Check if validator already exists
	// This is necessary because Set() with the same primary key (power, consensus_address)
	// will overwrite without checking unique indexes, allowing operator address to change
	exists, err := k.HasValidator(ctx, consAddress)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrValidatorAlreadyExists
	}

	// Validate operator and consensus pubkey are different
	if err := k.ValidateOperatorAndConsensusPubKeyDifferent(validator.Metadata.OperatorAddress, validator.PubKey); err != nil {
		return err
	}

	// Create validator with initial power (from validator.Power)
	// Keepers are the only ones that can set power to > 0.
	// Thus, on creation, we queue an update for consensus.
	if validator.Power > 0 {
		// Checkpoint all validators if requested. The only time we don't want to do this is
		// during ImportGenesis, where we re-add all the validators one by one.
		if checkpoint {
			if err := k.CheckpointAllValidators(ctx); err != nil {
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

		// Adjust total power
		if err := k.AdjustTotalPower(ctx, validator.Power); err != nil {
			return err
		}
	}

	// The unique index on consensus address will prevent duplicates when
	// primary key differs, but we also need the HasValidator check above
	// to prevent overwrites with the same primary key
	key := collections.Join(validator.Power, consAddress.String())
	return k.validators.Set(ctx, key, validator)
}

// HasValidator checks if a validator exists by consensus address.
func (k *Keeper) HasValidator(ctx sdk.Context, consAddress sdk.ConsAddress) (bool, error) {
	// Use ConsensusAddress index to check if validator exists
	_, err := k.validators.Indexes.ConsensusAddress.MatchExact(ctx, consAddress.String())
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetValidator retrieves a validator by consensus address.
func (k *Keeper) GetValidator(ctx sdk.Context, consAddress sdk.ConsAddress) (types.Validator, error) {
	// Use ConsensusAddress index to find the composite key
	compositeKey, err := k.validators.Indexes.ConsensusAddress.MatchExact(ctx, consAddress.String())
	if err != nil {
		return types.Validator{}, err
	}
	// Get the full validator using the composite key
	return k.validators.Get(ctx, compositeKey)
}

// createABCIValidatorUpdate creates a CometBFT validator update from a validator's public key and power.
func (k *Keeper) createABCIValidatorUpdate(pubKeyAny *codectypes.Any, power int64) (abci.ValidatorUpdate, error) {
	var pubKeySDK cryptotypes.PubKey
	if err := k.cdc.UnpackAny(pubKeyAny, &pubKeySDK); err != nil {
		return abci.ValidatorUpdate{}, err
	}

	pubKeyCMT, err := codec.ToCmtProtoPublicKey(pubKeySDK)
	if err != nil {
		return abci.ValidatorUpdate{}, err
	}

	return abci.ValidatorUpdate{PubKey: pubKeyCMT, Power: power}, nil
}

// SetValidatorPower sets the power for a validator and handles the rekeying of the primary power key.
// This checkpoints all validators before making the change to ensure accurate fee distribution.
func (k *Keeper) SetValidatorPower(ctx sdk.Context, consAddress sdk.ConsAddress, power int64) error {
	// Checkpoint all validators before any power change
	if err := k.CheckpointAllValidators(ctx); err != nil {
		return err
	}

	// Get existing validator and old composite key
	oldCompositeKey, err := k.validators.Indexes.ConsensusAddress.MatchExact(ctx, consAddress.String())
	if err != nil {
		return err
	}

	validator, err := k.validators.Get(ctx, oldCompositeKey)
	if err != nil {
		return err
	}

	// Store old power for delta calculation
	oldPower := validator.Power
	delta := power - oldPower

	if err := k.AdjustTotalPower(ctx, delta); err != nil {
		return err
	}

	// Delete old entry
	if err := k.validators.Remove(ctx, oldCompositeKey); err != nil {
		return err
	}

	// Update power and insert with new key
	validator.Power = power
	newKey := collections.Join(power, consAddress.String())
	return k.validators.Set(ctx, newKey, validator)
}

// GetValidatorPower gets the power for a validator by consensus address.
func (k *Keeper) GetValidatorPower(ctx sdk.Context, consAddress sdk.ConsAddress) (int64, error) {
	// The power is in the composite key itself
	compositeKey, err := k.validators.Indexes.ConsensusAddress.MatchExact(ctx, consAddress.String())
	if err != nil {
		return 0, err
	}
	return compositeKey.K1(), nil
}

// GetValidatorByConsAddress retrieves a validator by consensus address.
func (k *Keeper) GetValidatorByConsAddress(ctx sdk.Context, consAddr sdk.ConsAddress) (types.Validator, error) {
	// Use consensus address index to find composite key
	compositeKey, err := k.validators.Indexes.ConsensusAddress.MatchExact(ctx, consAddr.String())
	if err != nil {
		return types.Validator{}, err
	}
	return k.validators.Get(ctx, compositeKey)
}

// GetValidatorByOperatorAddress retrieves a validator by operator address using the secondary index.
func (k *Keeper) GetValidatorByOperatorAddress(ctx sdk.Context, operatorAddr sdk.AccAddress) (types.Validator, error) {
	// Use operator address index to find composite key
	compositeKey, err := k.validators.Indexes.OperatorAddress.MatchExact(ctx, operatorAddr.String())
	if err != nil {
		return types.Validator{}, err
	}
	return k.validators.Get(ctx, compositeKey)
}

// IterateValidators iterates over validators with a custom range and callback function.
func (k *Keeper) IterateValidators(ctx sdk.Context, ranger *collections.Range[collections.Pair[int64, string]], callback func(power int64, validator types.Validator) (stop bool, err error)) error {
	return k.validators.Walk(ctx, ranger, func(key collections.Pair[int64, string], val types.Validator) (bool, error) {
		power := key.K1()
		return callback(power, val)
	})
}

// GetTotalPower returns the total validator power across all active validators.
func (k *Keeper) GetTotalPower(ctx sdk.Context) (int64, error) {
	power, err := k.totalPower.Get(ctx)
	if err != nil {
		if ctx.BlockHeight() == 0 {
			// The only context where we don't want to send an error is on genesis.
			// Total power should never be 0 otherwise.
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

// GetAllValidators returns all validators in the store.
// Validators are returned in descending power order.
func (k *Keeper) GetAllValidators(ctx sdk.Context) ([]types.Validator, error) {
	var validators []types.Validator
	// Iterate validators in descending power order
	ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()
	err := k.validators.Walk(ctx, ranger, func(key collections.Pair[int64, string], val types.Validator) (bool, error) {
		validators = append(validators, val)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return validators, nil
}
