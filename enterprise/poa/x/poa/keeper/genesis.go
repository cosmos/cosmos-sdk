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
	abci "github.com/cometbft/cometbft/abci/types"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	v3 "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/migrations/v3"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExportGenesis exports the current state of the keeper as a GenesisState.
// It retrieves all parameters and validators from the store and returns them
// in a format suitable for genesis export.
func (k *Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	// Export per-validator allocated fees
	var allocatedFees []types.GenesisAllocatedFees
	err = k.validatorAllocatedFees.Walk(ctx, nil, func(consAddr string, fees types.ValidatorFees) (bool, error) {
		if !fees.Fees.IsZero() {
			allocatedFees = append(allocatedFees, types.GenesisAllocatedFees{
				ConsensusAddress: consAddr,
				Fees:             fees.Fees,
			})
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		Params:        params,
		Validators:    validators,
		AllocatedFees: allocatedFees,
	}, nil
}

// InitGenesis initializes the keeper state from a GenesisState.
// It sets the module parameters and creates all validators specified in the genesis.
// Returns the validator updates that should be applied to the consensus engine.
func (k *Keeper) InitGenesis(ctx sdk.Context, cdc codec.BinaryCodec, genesis *types.GenesisState) ([]abci.ValidatorUpdate, error) {
	// Set module parameters
	if err := k.UpdateParams(ctx, genesis.Params); err != nil {
		return nil, err
	}

	// Create all validators from genesis
	for _, validator := range genesis.Validators {
		var pubKey cryptotypes.PubKey
		if err := cdc.UnpackAny(validator.PubKey, &pubKey); err != nil {
			return nil, err
		}

		consAddress := sdk.GetConsAddress(pubKey)
		if err := k.CreateValidator(ctx, consAddress, validator, false); err != nil {
			return nil, err
		}
	}

	// Restore per-validator allocated fees and recompute totalAllocatedFees
	var totalAllocated sdk.DecCoins
	for _, entry := range genesis.AllocatedFees {
		if err := k.validatorAllocatedFees.Set(ctx, entry.ConsensusAddress, types.ValidatorFees{Fees: entry.Fees}); err != nil {
			return nil, err
		}
		totalAllocated = totalAllocated.Add(entry.Fees...)
	}
	if !totalAllocated.IsZero() {
		if err := k.totalAllocatedFees.Set(ctx, types.ValidatorFees{Fees: totalAllocated}); err != nil {
			return nil, err
		}
	}

	// Exports from poa v1.x can carry rounding dust above the module balance;
	// trim it the way the v3 migration does. v1.x rounded each validator's
	// share to nearest, overshooting by less than 1e-18 per validator, and the
	// overshoot did not build up across checkpoints. Validators are never
	// deleted, only re-keyed, so the dust is at most 1e-18 per exported
	// validator per denom. A larger excess is not rounding: the genesis is
	// inconsistent, or bank's genesis has not run before poa's, so the balance
	// reads as empty.
	moduleAddr := k.authKeeper.GetModuleAddress(types.ModuleName)
	balance := sdk.NewDecCoinsFromCoins(k.bankKeeper.GetAllBalances(ctx, moduleAddr)...)
	excess := totalAllocated.Sub(totalAllocated.Intersect(balance))
	maxDust := math.LegacyNewDecWithPrec(int64(len(genesis.Validators)), math.LegacyPrecision)
	for _, coin := range excess {
		if coin.Amount.GT(maxDust) {
			return nil, sdkerrors.Wrapf(
				types.ErrInvalidAllocatedFees,
				"genesis allocated fees %s exceed module balance %s by %s, more than the rounding dust limit of %s per denom; bank genesis must run before poa",
				totalAllocated, balance, excess, maxDust,
			)
		}
	}
	if !excess.IsZero() {
		if err := v3.Migrate(ctx, balance, k.validatorAllocatedFees, k.totalAllocatedFees); err != nil {
			return nil, err
		}
	}

	// Compute the initial changeset and seed LastCommittedPower
	updates, err := k.ReapValidatorUpdates(ctx)
	if err != nil {
		return nil, err
	}
	if err := k.SetLastCommittedPower(ctx, updates); err != nil {
		return nil, err
	}
	return updates, nil
}
