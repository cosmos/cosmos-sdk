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
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
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

	return &types.GenesisState{
		Params:     params,
		Validators: validators,
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

	// Return queued validator updates
	updates := k.ReapValidatorUpdates(ctx)

	return updates, nil
}
