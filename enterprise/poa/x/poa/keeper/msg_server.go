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
	"strconv"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

// MsgServer implements the module's gRPC message service.
type MsgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*MsgServer)(nil)

// NewMsgServer creates a new MsgServer instance.
func NewMsgServer(keeper *Keeper) *MsgServer {
	return &MsgServer{
		keeper: keeper,
	}
}

// UpdateParams updates the module parameters. Only the admin can update parameters.
func (s *MsgServer) UpdateParams(
	ctx context.Context, req *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := req.Validate(s.keeper.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	admin, err := s.keeper.Admin(sdkCtx)
	if err != nil {
		return nil, err
	}

	if req.Admin != admin {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", admin, req.Admin)
	}

	if err := s.keeper.UpdateParams(sdkCtx, req.Params); err != nil {
		return nil, err
	}

	// Emit event for parameter update
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateParams,
			sdk.NewAttribute(types.AttributeKeyAdmin, req.Admin),
			sdk.NewAttribute(types.AttributeKeyParams, req.Params.String()),
		),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}

// CreateValidator creates a new validator with zero power. The validator must be activated by the admin.
func (s *MsgServer) CreateValidator(
	ctx context.Context, req *types.MsgCreateValidator,
) (*types.MsgCreateValidatorResponse, error) {
	var pubKey cryptotypes.PubKey
	if err := s.keeper.cdc.UnpackAny(req.PubKey, &pubKey); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := s.keeper.validatePubkeyType(sdkCtx, pubKey); err != nil {
		return nil, err
	}

	if err := req.Validate(s.keeper.authKeeper.AddressCodec()); err != nil {
		return nil, err
	}

	consAddress := sdk.GetConsAddress(pubKey)

	// Prevent using the same key for both operator and consensus
	if err := s.keeper.ValidateOperatorAndConsensusPubKeyDifferent(req.OperatorAddress, req.PubKey); err != nil {
		return nil, err
	}

	validator := types.Validator{
		PubKey: req.PubKey,
		Power:  0, // Validators are created with 0 power
		Metadata: &types.ValidatorMetadata{
			Moniker:         req.Moniker,
			Description:     req.Description,
			OperatorAddress: req.OperatorAddress,
		},
	}

	if err := s.keeper.CreateValidator(sdkCtx, consAddress, validator, true); err != nil {
		return nil, err
	}

	// Emit event for validator creation
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, req.OperatorAddress),
			sdk.NewAttribute(types.AttributeKeyConsensusAddress, consAddress.String()),
			sdk.NewAttribute(types.AttributeKeyMoniker, req.Moniker),
			sdk.NewAttribute(types.AttributeKeyPower, "0"),
		),
	)

	return &types.MsgCreateValidatorResponse{}, nil
}

// UpdateValidators updates one or more validators. Only the admin can update validators.
func (s *MsgServer) UpdateValidators(
	ctx context.Context, req *types.MsgUpdateValidators,
) (*types.MsgUpdateValidatorsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	admin, err := s.keeper.Admin(sdkCtx)
	if err != nil {
		return nil, err
	}

	if req.Admin != admin {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", admin, req.Admin)
	}

	if err := s.keeper.UpdateValidators(sdkCtx, req.Validators); err != nil {
		return nil, err
	}

	// Emit events for each validator update
	for _, validator := range req.Validators {
		var pubKey cryptotypes.PubKey
		if err := s.keeper.cdc.UnpackAny(validator.PubKey, &pubKey); err != nil {
			return nil, err
		}
		consAddress := sdk.GetConsAddress(pubKey)

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUpdateValidator,
				sdk.NewAttribute(types.AttributeKeyOperatorAddress, validator.Metadata.OperatorAddress),
				sdk.NewAttribute(types.AttributeKeyConsensusAddress, consAddress.String()),
				sdk.NewAttribute(types.AttributeKeyPower, strconv.FormatInt(validator.Power, 10)),
			),
		)
	}

	return &types.MsgUpdateValidatorsResponse{}, nil
}

// WithdrawFees allows a validator operator to withdraw their accumulated transaction fees.
func (s *MsgServer) WithdrawFees(
	ctx context.Context, req *types.MsgWithdrawFees,
) (*types.MsgWithdrawFeesResponse, error) {
	operatorAddress, err := sdk.AccAddressFromBech32(req.Operator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid validator address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	coins, err := s.keeper.WithdrawValidatorFees(sdkCtx, operatorAddress)
	if err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawFees,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, req.Operator),
			sdk.NewAttribute(types.AttributeKeyAmount, coins.String()),
		),
	)

	return &types.MsgWithdrawFeesResponse{}, nil
}
