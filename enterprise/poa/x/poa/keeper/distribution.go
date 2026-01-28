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
// See ./enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

// CheckpointAllValidators allocates pending fees to all validators.
// This must be called every time validator power changes or a validator
// withdraws its fees.
func (k *Keeper) CheckpointAllValidators(ctx sdk.Context) error {
	// Get unallocated fees
	unallocated, err := k.getUnallocatedFees(ctx)
	if err != nil {
		return err
	}

	// Get total power
	totalPower, err := k.GetTotalPower(ctx)
	if err != nil {
		return err
	}

	// If no unallocated fees or no power, nothing to checkpoint
	if unallocated.IsZero() {
		return nil
	}

	// Iterate validators in descending power order
	ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()
	err = k.validators.Walk(ctx, ranger, func(key collections.Pair[int64, string], validator types.Validator) (bool, error) {
		power := key.K1()

		// Stop iteration when we reach validators with power 0
		if power == 0 {
			return true, nil
		}

		// Calculate this validator's share using the shared helper
		validatorPendingFees := calculateValidatorPendingFees(power, totalPower, unallocated)

		// Update validator accumulated fees
		validator.AllocatedFees = validator.AllocatedFees.Add(validatorPendingFees...)

		// Update total allocated
		if err := k.adjustTotalAllocated(ctx, validatorPendingFees); err != nil {
			return true, err
		}

		return false, k.validators.Set(ctx, key, validator)
	})
	if err != nil {
		ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName)).Debug("error checkpointing all validator fees", "error", err)
		return err
	}

	ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName)).Info("checkpointed all validator fees", "total", totalPower, "unallocated", unallocated)
	return nil
}

// getUnallocatedFees returns the unallocated fees
// Returns zero values if there are no unallocated fees or no validators.
func (k *Keeper) getUnallocatedFees(ctx sdk.Context) (unallocated sdk.DecCoins, err error) {
	// Get fee collector balance
	feeCollector := k.authKeeper.GetModuleAccount(ctx, authtypes.FeeCollectorName)
	feeCollectorBalance := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())

	// If no fees in collector, return zero
	if feeCollectorBalance.IsZero() {
		return sdk.DecCoins{}, nil
	}

	feeCollectorBalanceDec := sdk.NewDecCoinsFromCoins(feeCollectorBalance...)

	// Get total allocated
	totalAllocated, err := k.getTotalAllocated(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate unallocated fees = fee_collector - total_allocated
	unallocated = feeCollectorBalanceDec.Sub(totalAllocated)

	// If no unallocated fees, return zero
	if unallocated.IsZero() || !unallocated.IsAllPositive() {
		return sdk.DecCoins{}, nil
	}

	return unallocated, nil
}

// calculateValidatorPendingFees calculates a validator's share of unallocated fees.
// Formula: unallocated * validator_power / total_power.
// This is a helper function used by both CheckpointAllValidators and WithdrawableFees query.
// Panics if totalPower == 0, as this indicates an invalid state.
func calculateValidatorPendingFees(validatorPower, totalPower int64, unallocated sdk.DecCoins) sdk.DecCoins {
	if totalPower == 0 {
		panic("totalPower cannot be zero when calculating validator pending fees")
	}

	if validatorPower == 0 {
		return sdk.DecCoins{}
	}

	totalPowerDec := math.LegacyNewDec(totalPower)
	validatorPowerDec := math.LegacyNewDec(validatorPower)
	return unallocated.MulDec(validatorPowerDec).QuoDec(totalPowerDec)
}

// WithdrawValidatorFees withdraws accumulated fees for a validator
// Returns the amount withdrawn as coins.
func (k *Keeper) WithdrawValidatorFees(ctx sdk.Context, validatorAddr sdk.AccAddress) (sdk.Coins, error) {
	compositeKey, err := k.validators.Indexes.OperatorAddress.MatchExact(ctx, validatorAddr.String())
	if err != nil {
		return nil, err
	}

	// Checkpoint all validators to allocate pending fees before withdrawal
	if err := k.CheckpointAllValidators(ctx); err != nil {
		return nil, err
	}

	validator, err := k.validators.Get(ctx, compositeKey)
	if err != nil {
		return nil, err
	}

	// Truncate DecCoins to Coins, preserving the decimal remainder
	coins, remainder := validator.AllocatedFees.TruncateDecimal()

	// If no fees to withdraw, return early
	if coins.IsZero() {
		return coins, nil
	}

	// Transfer fees from fee collector to validator address
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, validatorAddr, coins)
	if err != nil {
		return nil, err
	}

	// Subtract withdrawn coins from total allocated
	withdrawnDec := sdk.NewDecCoinsFromCoins(coins...)
	if err := k.adjustTotalAllocated(ctx, withdrawnDec.MulDec(math.LegacyNewDec(-1))); err != nil {
		return nil, err
	}

	// Update validator with the decimal remainder (prevents dust accumulation)
	validator.AllocatedFees = remainder
	err = k.validators.Set(ctx, compositeKey, validator)
	if err != nil {
		return nil, err
	}

	return coins, nil
}

// getValidatorAllocatedFees returns the accumulated fees for a validator
func (k *Keeper) getValidatorAllocatedFees(ctx sdk.Context, consAddr sdk.ConsAddress) (sdk.DecCoins, error) {
	validator, err := k.GetValidator(ctx, consAddr)
	if err != nil {
		return nil, err
	}
	return validator.AllocatedFees, nil
}

// getTotalAllocated returns the total allocated fees across all validators
func (k *Keeper) getTotalAllocated(ctx sdk.Context) (sdk.DecCoins, error) {
	allocated, err := k.totalAllocatedFees.Get(ctx)
	if err != nil {
		// If not found, return empty (happens on genesis)
		return sdk.DecCoins{}, nil
	}
	return allocated.Fees, nil
}

// adjustTotalAllocated adjusts the total allocated by the given delta.
// This is an internal helper function to track the sum of all allocated fees across validators.
// Returns an error if the adjustment would result in a negative total.
func (k *Keeper) adjustTotalAllocated(ctx sdk.Context, delta sdk.DecCoins) error {
	current, err := k.getTotalAllocated(ctx)
	if err != nil {
		return err
	}

	newTotal := current.Add(delta...)

	// Prevent total allocated from going below zero
	if !newTotal.IsAllPositive() && !newTotal.IsZero() {
		return fmt.Errorf("cannot adjust total allocated below zero: current=%s, delta=%s, result=%s", current, delta, newTotal)
	}

	return k.totalAllocatedFees.Set(ctx, types.ValidatorFees{Fees: newTotal})
}
