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
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

// Params returns the current module parameters.
func (k *Keeper) Params(
	ctx context.Context, _ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := k.GetParams(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Validator returns a validator by consensus address or operator address.
func (k *Keeper) Validator(
	ctx context.Context, req *types.QueryValidatorRequest,
) (*types.QueryValidatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Try parsing as consensus address first
	consAddress, consErr := sdk.ConsAddressFromBech32(req.Address)
	if consErr == nil {
		// Successfully parsed as consensus address
		validator, err := k.GetValidator(sdkCtx, consAddress)
		if err != nil {
			return nil, err
		}
		return &types.QueryValidatorResponse{
			Validator: validator,
		}, nil
	}

	// Try parsing as operator (account) address
	operatorAddress, opErr := sdk.AccAddressFromBech32(req.Address)
	if opErr == nil {
		// Successfully parsed as operator address
		validator, err := k.GetValidatorByOperatorAddress(sdkCtx, operatorAddress)
		if err != nil {
			return nil, err
		}
		return &types.QueryValidatorResponse{
			Validator: validator,
		}, nil
	}

	// If both failed, return an error indicating invalid address format
	return nil, fmt.Errorf(
		"address must be either a valid consensus address or operator address: cons_err=%w, op_err=%w",
		consErr, opErr,
	)
}

// Validators queries all validators with pagination
// Behavior is to return validators in descending order by power. The pagination reverse field is ignored.
func (k *Keeper) Validators(
	ctx context.Context, req *types.QueryValidatorsRequest,
) (*types.QueryValidatorsResponse, error) {
	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{}
	}

	// fix reverse to true so we always return in descending order
	pageReq.Reverse = true

	validators, pageRes, err := query.CollectionPaginate(
		ctx,
		k.validators,
		pageReq,
		func(key collections.Pair[int64, string], value types.Validator) (types.Validator, error) {
			return value, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

// WithdrawableFees returns the total withdrawable fees for a validator (allocated + pending).
func (k *Keeper) WithdrawableFees(
	ctx context.Context, req *types.QueryWithdrawableFeesRequest,
) (*types.QueryWithdrawableFeesResponse, error) {
	operatorAddress, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Use the secondary index to find the composite key
	compositeKey, err := k.validators.Indexes.OperatorAddress.MatchExact(sdkCtx, operatorAddress.String())
	if err != nil {
		return nil, err
	}

	// Get the validator
	validator, err := k.validators.Get(sdkCtx, compositeKey)
	if err != nil {
		return nil, err
	}

	// Calculate pending fees using lazy distribution formula:
	// allocated + validator_power * (fee_collector - total_allocated) / total_power
	totalFees := validator.AllocatedFees

	// Get total power
	totalPower, err := k.GetTotalPower(sdkCtx)
	if err != nil {
		return nil, err
	}

	// Get unallocated fees
	unallocated, err := k.getUnallocatedFees(sdkCtx)
	if err != nil {
		return nil, err
	}

	// If there are unallocated fees, calculate this validator's pending share
	if !unallocated.IsZero() {
		// Calculate pending fees using the shared helper
		pendingFees := calculateValidatorPendingFees(validator.Power, totalPower, unallocated)
		totalFees = totalFees.Add(pendingFees...)
	}

	return &types.QueryWithdrawableFeesResponse{
		Fees: types.ValidatorFees{Fees: totalFees},
	}, nil
}

// TotalPower returns the total voting power of all active validators.
func (k *Keeper) TotalPower(
	ctx context.Context, _ *types.QueryTotalPowerRequest,
) (*types.QueryTotalPowerResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	totalPower, err := k.GetTotalPower(sdkCtx)
	if err != nil {
		return nil, err
	}

	return &types.QueryTotalPowerResponse{
		TotalPower: totalPower,
	}, nil
}
